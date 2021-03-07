using Newtonsoft.Json;
using System.Collections.Generic;

namespace ota_update_management
{
    public class Manifest
    {
        [JsonProperty("configId")]
        public string ConfigId {get; set;}

        [JsonProperty("baselineId")]
        public string BaselineId {get; set;}
        
        [JsonProperty("appsettings")]
        public Dictionary<string, SwMetadata> AppSettings {get; set;}
    }

    public class SwMetadata
    {
        public string ImageName {get; set;}

        public string Version {get; set;}
    }
}