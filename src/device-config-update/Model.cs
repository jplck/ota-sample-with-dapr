using Newtonsoft.Json;
using System.Collections.Generic;
using System.ComponentModel;

namespace Cloud.DeviceConfiguration
{
    public class DeviceSoftwareDefinition
    {
        [JsonProperty("configId")]
        public string ConfigId {get; set;}

        [JsonProperty("baselineId")]
        public string BaselineId {get; set;}

        [JsonProperty("priority", DefaultValueHandling = DefaultValueHandling.Populate)]
        [DefaultValue(10)]
        public int Priority {get; set;}
        
        [JsonProperty("manifest")]
        public Manifest Manifest {get; set;}
    }

    public class Manifest
    {
        [JsonProperty("description")]
        public string Description {get; set;}

        [JsonProperty("packages")]
        public Dictionary<string, SwMetadata> Packages {get; set;}
    }

    public class SwMetadata
    {
        [JsonProperty("imageName")]
        public string ImageName {get; set;}

        [JsonProperty("version")]
        public string Version {get; set;}
    }
}