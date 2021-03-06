using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;
using Dapr.Client;
using System.Linq;
using Microsoft.Azure.Devices;
using Newtonsoft.Json;

namespace Cloud.DeviceConfiguration.Controllers
{
    [ApiController]
    [Route("/manifest")]
    public class UpdateDeviceConfigurationController : ControllerBase
    {
        private readonly ILogger<UpdateDeviceConfigurationController> _logger;

        private readonly DaprClient _daprClient;

        private RegistryManager _registryManager;

        private const string MANIFEST_PROP_NAME = "properties.desired.devicesoftwaredefinition";

        public UpdateDeviceConfigurationController(DaprClient daprClient, ILogger<UpdateDeviceConfigurationController> logger)
        {
            _logger = logger;
            _daprClient = daprClient;
        }

        public async Task<RegistryManager> GetRegistryManagerAsync()
        {
            if (_registryManager is not null) { return _registryManager; }

            var hubConnectionSecret = await _daprClient.GetSecretAsync("secretstore", "IotHubConnectionString");
            var hubConnectionString = hubConnectionSecret.Values.First<string>();

            _registryManager = RegistryManager.CreateFromConnectionString(hubConnectionString);
            return _registryManager;
        }

        private void Verifiy(DeviceSoftwareDefinition message)
        {
            _ = message.ConfigId ?? throw new ArgumentNullException();
            _ = message.Manifest ?? throw new ArgumentNullException();
        }

        [HttpPatch]
        [Consumes("application/json")]
        public async Task<ActionResult<string>> UpdateManifest([FromBody] DeviceSoftwareDefinition message)
        {
            try
            {
                Verifiy(message);

                var registryManager = await GetRegistryManagerAsync();

                _logger.LogInformation($"Received manifest with config id {message.ConfigId}");

                var configId = message.BaselineId is null ?
                     message.ConfigId : 
                     message.BaselineId;

                var existingConfig = await _registryManager.GetConfigurationAsync(configId);

                var existingManifest = JsonConvert.DeserializeObject<Manifest>(
                    existingConfig.Content.DeviceContent[MANIFEST_PROP_NAME].ToString()
                );

                var existingPackages = existingManifest.Packages;
                var receivedPackages = message.Manifest.Packages;

                _logger.LogInformation("Creating config.");

                if (message.BaselineId == message.ConfigId || message.BaselineId is null) {
                    _logger.LogInformation("Removing existing configuration. (essentially an override)");
                    await registryManager.RemoveConfigurationAsync(message.ConfigId);
                }

                receivedPackages.ToList().ForEach(x => existingPackages[x.Key] = x.Value);

                var updatedManifest = message.Manifest;
                updatedManifest.Packages = existingPackages;

                var priority = existingConfig.Priority < message.Priority ? message.Priority : existingConfig.Priority + 1;

                var config = GenerateConfiguration(
                    message.ConfigId, 
                    existingConfig.TargetCondition, 
                    priority, 
                    new Dictionary<string, object>() {
                    { 
                        MANIFEST_PROP_NAME,
                        updatedManifest
                    }
                });
                
                await _registryManager.AddConfigurationAsync(config);

                return new OkObjectResult(existingConfig.ToString());

            }
            catch (Exception ex)
            {
                _logger.LogError(ex.Message);
                return new BadRequestObjectResult("Unable to execute config update.");
            }
        }

        [HttpPost]
        [Consumes("application/json")]
        public async Task<ActionResult<string>> CreateManifest([FromBody] DeviceSoftwareDefinition message, [FromServices] DaprClient daprClient)
        {
            try
            {
                Verifiy(message);

                var registryManager = await GetRegistryManagerAsync();

                _logger.LogInformation($"Received manifest with config id {message.ConfigId}");

                var config = GenerateConfiguration(message.ConfigId, "*", message.Priority, new Dictionary<string, object>() {
                    { 
                        MANIFEST_PROP_NAME,
                        message.Manifest
                    }
                });

                await registryManager.AddConfigurationAsync(config);
                
                return new OkObjectResult(config.ToString());

            }
            catch (Exception ex)
            {
                _logger.LogError(ex.Message);
                return new BadRequestObjectResult("Unable to execute config update.");
            }
        }

        private Configuration GenerateConfiguration(
            string configId, 
            string targetCondition, 
            int priority, 
            Dictionary<string, object> deviceContent
        )
        {
            return new Configuration(configId)
            {
                TargetCondition = targetCondition,
                Priority = priority,
                Content = new ConfigurationContent()
                {
                    DeviceContent = deviceContent
                }
            };
        }
    }
}