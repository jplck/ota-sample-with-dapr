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

        [Topic("telemetryeventhub", "daprhub1")]
        [HttpPost("ehtelemetry")]
        public IActionResult EhTelemetryReceive()
        {
            _logger.LogInformation("test");
            return new OkResult();
        }

        [Topic("servicebus-pubsub", "deviceevents")]
        [HttpPost("sbtelemetry")]
        public IActionResult SbTelemetryReceive()
        {
            _logger.LogInformation("test");
            return new OkResult();
        }
    }

}