using Newtonsoft.Json;
using System.Collections.Generic;
using Microsoft.Azure.Devices;

namespace ota_update_management
{
    public class DeviceSoftwareDefinition
    {
        [JsonProperty("configId")]
        public string ConfigId {get; set;}

        [JsonProperty("baselineId")]
        public string BaselineId {get; set;}
        
        [JsonProperty("manifest")]
        public Manifest Manifest {get; set;}
    }

    public class Manifest
    {
        [JsonProperty("description")]
        public string Description {get; set;}

        [JsonProperty("definitions")]
        public Dictionary<string, SwMetadata> Definitions {get; set;}
    }

    public class SwMetadata
    {
        [JsonProperty("imageName")]
        public string ImageName {get; set;}

        [JsonProperty("version")]
        public string Version {get; set;}
    }
}