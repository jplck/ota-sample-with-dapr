using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;
using Dapr.Client;
using Microsoft.Azure.Devices;
using Newtonsoft.Json;

namespace ota_update_management.Controllers
{
    [ApiController]
    [Route("[controller]")]
    public class UpdateManifestController : ControllerBase
    {
        private readonly ILogger<UpdateManifestController> _logger;
        private const string APP_SETTINGS_PROP_NAME = "properties.desired.appsettings";

        public UpdateManifestController(ILogger<UpdateManifestController> logger)
        {
            _logger = logger;
        }

        [HttpPost]
        [Consumes("application/json")]
        public async Task<ActionResult<string>> Update([FromBody] Manifest message, [FromServices] DaprClient daprClient)
        {
            try
            {
                _ = message.ConfigId ?? throw new ArgumentNullException();
                _ = message.AppSettings ?? throw new ArgumentNullException();

                _logger.LogInformation($"Received manifest with config id {message.ConfigId}");

                var secretStore = "azurekeyvault";

                #if (DEBUG) 
                    secretStore = "localsecretstore";
                #endif

                var hubConnectionSecret = await daprClient.GetSecretAsync(secretStore, "IotHubConnectionString");
                var hubConnectionString = hubConnectionSecret.Values.First<string>();

                var registryManager = RegistryManager.CreateFromConnectionString(hubConnectionString);

                var baselineConfig = message.BaselineId is not null ? await registryManager.GetConfigurationAsync(message.BaselineId) : null;
                var rawBaselineAppSettingsString = baselineConfig?.Content?.DeviceContent?[APP_SETTINGS_PROP_NAME]?.ToString();

                var config = new Configuration(message.ConfigId)
                {
                    Priority = baselineConfig?.Priority ?? 0,
                    TargetCondition = "*",
                    Content = new ConfigurationContent()
                    {
                        DeviceContent = new Dictionary<string, object>()
                            {
                                { 
                                    APP_SETTINGS_PROP_NAME,
                                     rawBaselineAppSettingsString is null ? 
                                        new Dictionary<string, SwMetadata>() : 
                                        JsonConvert.DeserializeObject<IDictionary<string, SwMetadata>>(rawBaselineAppSettingsString)
                                }
                            }
                    }
                };

                var appSettings = (IDictionary<string, SwMetadata>)config.Content.DeviceContent[APP_SETTINGS_PROP_NAME];

                foreach (KeyValuePair<string, SwMetadata> setting in message.AppSettings)
                {
                    if (appSettings.ContainsKey(setting.Key))
                    {
                        appSettings[setting.Key] = setting.Value;
                        continue;
                    }
                    appSettings.Add(setting.Key, setting.Value);
                }

                config.Priority += 1;
                config.Content.DeviceContent[APP_SETTINGS_PROP_NAME] = appSettings;

                _logger.LogInformation("Creating config.");

                if (message.BaselineId == message.ConfigId) {
                    _logger.LogInformation("Removing existing base configuration. (ConfigId == BaseId)");
                    await registryManager.RemoveConfigurationAsync(message.BaselineId);
                }

                await registryManager.AddConfigurationAsync(config);
                
                return new OkObjectResult(config.ToString());

            }
            catch (Exception ex)
            {
                _logger.LogError(ex.Message);
                return new BadRequestObjectResult("Unable to execute config update.");
            }
        }
    }
}