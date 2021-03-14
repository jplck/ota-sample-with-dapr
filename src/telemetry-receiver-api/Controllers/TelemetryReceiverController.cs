using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Extensions.Logging;
using Dapr;
using Dapr.Client;

namespace telemetry_receiver_api.Controllers
{
    public class TelemetryReceiverController : ControllerBase
    {
        private readonly ILogger<TelemetryReceiverController> _logger;

        public TelemetryReceiverController(ILogger<TelemetryReceiverController> logger)
        {
            _logger = logger;
        }

        [HttpPost("telemetryeventhub")]
        public ActionResult<string> EhTelemetryReceive([FromBody] dynamic data)
        {
            return new OkObjectResult("OK");
        }
/*
        [Topic("servicebus-pubsub", "deviceevents")]
        [HttpPost("deviceevents")]
        public ActionResult<string> SbTelemetryReceive([FromBody] string data, [FromServices] DaprClient dapr)
        {
            _logger.LogInformation(data);
            return new OkObjectResult("OK");
        }*/
    }

}