using System;
using System.IO;
using UAssetAPI;
using UAssetAPI.UnrealTypes;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace UAssetBridge
{
    class Program
    {
        static void Main(string[] args)
        {
            if (args.Length < 2)
            {
                Console.WriteLine("Usage: UAssetBridge.exe <export/import> <folderPath>");
                Console.WriteLine("Commands:");
                Console.WriteLine("  export <folder> - Convert .uasset/.uexp files to .json");
                Console.WriteLine("  import <folder> - Convert .json files back to .uasset/.uexp");
                return;
            }

            string command = args[0].ToLower();
            string folderPath = args[1];

            if (!Directory.Exists(folderPath))
            {
                Console.WriteLine($"Error: Directory '{folderPath}' does not exist.");
                return;
            }

            try
            {
                switch (command)
                {
                    case "export":
                        ExportUAssets(folderPath);
                        break;
                    case "import":
                        ImportUAssets(folderPath);
                        break;
                    default:
                        Console.WriteLine($"Error: Unknown command '{command}'. Use 'export' or 'import'.");
                        break;
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Error: {ex.Message}");
                Console.WriteLine($"Stack trace: {ex.StackTrace}");
            }
        }

        static void ExportUAssets(string folderPath)
        {
            Console.WriteLine($"Starting export of UAsset files in: {folderPath}");

            var uassetFiles = Directory.GetFiles(folderPath, "*.uasset", SearchOption.AllDirectories);

            Console.WriteLine($"Found {uassetFiles.Length} .uasset files");

            int processed = 0;
            int failed = 0;
            int total = uassetFiles.Length;

            foreach (var file in uassetFiles)
            {
                try
                {
                    // Determine engine version (default to UE 5.3 for Grounded 2)
                    // UAssetAPI will try to auto-detect if this isn't quite right
                    EngineVersion engineVersion = EngineVersion.VER_UE5_3;

                    // Load the .uasset file
                    UAsset asset = new UAsset(file, engineVersion);

                    // Generate JSON path
                    string jsonPath = Path.ChangeExtension(file, ".json");

                    // Serialize to JSON
                    string json = asset.SerializeJson();

                    // Format the JSON for readability
                    JObject jsonObject = JObject.Parse(json);
                    string formattedJson = jsonObject.ToString(Formatting.Indented);

                    // Save formatted JSON
                    File.WriteAllText(jsonPath, formattedJson);

                    processed++;

                    if (processed % 10 == 0 || processed == total)
                    {
                        Console.WriteLine($"Progress: {processed}/{total} files processed");
                    }
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Error processing {file}: {ex.Message}");
                    failed++;
                }
            }

            Console.WriteLine($"Export completed. Processed {processed} files. {failed} failed.");
        }

        static void ImportUAssets(string folderPath)
        {
            Console.WriteLine($"Starting import of JSON files in: {folderPath}");

            var jsonFiles = Directory.GetFiles(folderPath, "*.json", SearchOption.AllDirectories);

            Console.WriteLine($"Found {jsonFiles.Length} .json files");

            int processed = 0;
            int failed = 0;
            int total = jsonFiles.Length;

            foreach (var file in jsonFiles)
            {
                try
                {
                    // Read JSON
                    string json = File.ReadAllText(file);

                    // Deserialize from JSON
                    UAsset asset = UAsset.DeserializeJson(json);

                    // Generate .uasset path
                    string uassetPath = Path.ChangeExtension(file, ".uasset");

                    // Write back to .uasset/.uexp
                    asset.Write(uassetPath);

                    processed++;

                    if (processed % 10 == 0 || processed == total)
                    {
                        Console.WriteLine($"Progress: {processed}/{total} files processed");
                    }
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"Error processing {file}: {ex.Message}");
                    failed++;
                }
            }

            Console.WriteLine($"Import completed. Processed {processed} files. {failed} failed.");
        }
    }
}
