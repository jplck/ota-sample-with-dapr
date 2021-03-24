using System;
using Dapr.Client;
using System.Threading.Tasks;

namespace telemetry_sender
{
    class Program
    {
        private const int _numberOfRuns = 1000;

        private const int _delay = 1000;
        private const string _ehBindingName = "telemetry-binding";

        static async Task Main(string[] args)
        {
            var daprClient = new DaprClientBuilder().Build();
            var idx = 0;
            while (idx <= _numberOfRuns)
            {
                await daprClient.InvokeBindingAsync(_ehBindingName, "create", idx);
                idx++;
                Console.WriteLine(idx);
                await Task.Delay(_delay);
            }
        }
    }
}
