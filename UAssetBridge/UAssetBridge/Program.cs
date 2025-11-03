using System;
using System.IO;
using UAssetAPI;
using UAssetAPI.UnrealTypes;
using UAssetAPI.Unversioned;
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
                Console.WriteLine("Usage: UAssetBridge.exe <export/import> <folderPath> [mappingsPath]");
                Console.WriteLine("Commands:");
                Console.WriteLine("  export <folder> [mappingsPath] - Convert .uasset/.uexp files to .json");
                Console.WriteLine("  import <folder> - Convert .json files back to .uasset/.uexp");
                Console.WriteLine();
                Console.WriteLine("  mappingsPath: Optional path to .usmap file for unversioned properties");
                return;
            }

            string command = args[0].ToLower();
            string folderPath = args[1];
            string? mappingsPath = args.Length > 2 ? args[2] : null;

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
                        ExportUAssets(folderPath, mappingsPath);
                        break;
                    case "import":
                        ImportUAssets(folderPath, mappingsPath);
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

        static void ExportUAssets(string folderPath, string? mappingsPath)
        {
            Console.WriteLine("=== UAssetBridge Export Starting ===");
            Console.WriteLine($"Received Asset Folder: {folderPath}");
            Console.WriteLine($"Received Mappings Path: {(string.IsNullOrEmpty(mappingsPath) ? "None" : mappingsPath)}");

            // Step 1: Create Usmap object from the provided file path
            // This is the critical step - we must pass a Usmap OBJECT, not a string path
            Usmap? mappings = null;
            if (!string.IsNullOrEmpty(mappingsPath))
            {
                if (!File.Exists(mappingsPath))
                {
                    Console.WriteLine($"⚠ ERROR: Mappings file not found at: {mappingsPath}");
                    Console.WriteLine("Continuing without mappings - export will be incomplete");
                }
                else
                {
                    try
                    {
                        Console.WriteLine($"Creating Usmap object from: {mappingsPath}");
                        mappings = new Usmap(mappingsPath);
                        Console.WriteLine($"✓ SUCCESS: Loaded {mappings.Schemas.Count} schemas from mappings file");
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"⚠ ERROR: Failed to create Usmap object: {ex.Message}");
                        Console.WriteLine($"Exception Type: {ex.GetType().Name}");
                        Console.WriteLine($"Stack trace: {ex.StackTrace}");
                        Console.WriteLine("Continuing without mappings - export will be incomplete");
                    }
                }
            }
            else
            {
                Console.WriteLine("⚠ No mappings file provided - export will be incomplete for unversioned properties");
            }

            var uassetFiles = Directory.GetFiles(folderPath, "*.uasset", SearchOption.AllDirectories);
            Console.WriteLine($"Found {uassetFiles.Length} .uasset file(s)");

            int processed = 0;
            int failed = 0;
            int total = uassetFiles.Length;

            foreach (var file in uassetFiles)
            {
                try
                {
                    Console.WriteLine($"\n--- Processing: {Path.GetFileName(file)} ---");

                    // Step 2: Parse engine version (using UE 5.4 for Grounded 2 v0.2.0.0)
                    EngineVersion engineVersion = EngineVersion.VER_UE5_4;
                    Console.WriteLine($"Engine Version: {engineVersion}");
                    Console.WriteLine($"Mappings Status: {(mappings != null ? $"Loaded ({mappings.Schemas.Count} schemas)" : "Not loaded")}");

                    // Step 3: Instantiate UAsset with the Usmap OBJECT (not string path)
                    Console.WriteLine("Creating UAsset object...");
                    UAsset asset = new UAsset(file, engineVersion, mappings);
                    Console.WriteLine("✓ UAsset object created successfully");

                    // Generate JSON path
                    string jsonPath = Path.ChangeExtension(file, ".json");

                    // Step 4: Serialize the fully-parsed asset to JSON
                    Console.WriteLine("Serializing to JSON...");
                    string json = asset.SerializeJson();
                    Console.WriteLine($"✓ JSON serialization complete ({json.Length} characters)");

                    // Format the JSON for readability
                    Console.WriteLine("Formatting JSON...");
                    JObject jsonObject = JObject.Parse(json);
                    string formattedJson = jsonObject.ToString(Formatting.Indented);
                    Console.WriteLine($"✓ JSON formatted ({formattedJson.Length} characters)");

                    // Step 5: Write the JSON string to the destination file
                    Console.WriteLine($"Writing to: {jsonPath}");
                    File.WriteAllText(jsonPath, formattedJson);
                    Console.WriteLine($"✓ File written successfully");

                    processed++;

                    if (processed % 10 == 0 || processed == total)
                    {
                        Console.WriteLine($"\nProgress: {processed}/{total} files processed");
                    }
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"⚠ ERROR processing {Path.GetFileName(file)}: {ex.Message}");
                    Console.WriteLine($"Exception Type: {ex.GetType().Name}");
                    Console.WriteLine($"Stack trace: {ex.StackTrace}");
                    failed++;
                }
            }

            Console.WriteLine($"\n=== Export Complete ===");
            Console.WriteLine($"Successfully processed: {processed} files");
            Console.WriteLine($"Failed: {failed} files");
        }

        static void ImportUAssets(string folderPath, string? mappingsPath = null)
        {
            Console.WriteLine("=== UAssetBridge Import Starting ===");
            Console.WriteLine($"Received JSON Folder: {folderPath}");
            Console.WriteLine($"Received Mappings Path: {(string.IsNullOrEmpty(mappingsPath) ? "None" : mappingsPath)}");

            // Load mappings if provided (needed for writing unversioned properties)
            Usmap? mappings = null;
            if (!string.IsNullOrEmpty(mappingsPath))
            {
                if (!File.Exists(mappingsPath))
                {
                    Console.WriteLine($"⚠ WARNING: Mappings file not found at: {mappingsPath}");
                    Console.WriteLine("Continuing without mappings - import may fail for unversioned assets");
                }
                else
                {
                    try
                    {
                        Console.WriteLine($"Loading Usmap object from: {mappingsPath}");
                        mappings = new Usmap(mappingsPath);
                        Console.WriteLine($"✓ SUCCESS: Loaded {mappings.Schemas.Count} schemas from mappings file");
                    }
                    catch (Exception ex)
                    {
                        Console.WriteLine($"⚠ ERROR: Failed to load Usmap object: {ex.Message}");
                        Console.WriteLine("Continuing without mappings - import may fail");
                    }
                }
            }
            else
            {
                Console.WriteLine("⚠ No mappings file provided - import may fail for unversioned properties");
            }

            var jsonFiles = Directory.GetFiles(folderPath, "*.json", SearchOption.AllDirectories);
            Console.WriteLine($"Found {jsonFiles.Length} .json file(s)");

            int processed = 0;
            int failed = 0;
            int total = jsonFiles.Length;

            foreach (var file in jsonFiles)
            {
                try
                {
                    Console.WriteLine($"\n--- Processing: {Path.GetFileName(file)} ---");

                    // Step 1: Read JSON file
                    Console.WriteLine("Reading JSON file...");
                    string json = File.ReadAllText(file);
                    Console.WriteLine($"✓ JSON file read ({json.Length} characters)");

                    // Step 2: Deserialize from JSON to UAsset object
                    Console.WriteLine("Deserializing JSON to UAsset...");
                    UAsset asset = UAsset.DeserializeJson(json);
                    Console.WriteLine("✓ UAsset object deserialized successfully");

                    // Step 3: Apply mappings to the deserialized asset (critical for writing)
                    if (mappings != null)
                    {
                        Console.WriteLine("Applying mappings to asset...");
                        asset.Mappings = mappings;
                        Console.WriteLine("✓ Mappings applied to asset");
                    }

                    // Step 4: Generate .uasset path
                    string uassetPath = Path.ChangeExtension(file, ".uasset");
                    Console.WriteLine($"Output path: {uassetPath}");

                    // Step 5: Write back to .uasset/.uexp files
                    Console.WriteLine("Writing to .uasset/.uexp files...");
                    asset.Write(uassetPath);
                    Console.WriteLine("✓ Files written successfully");

                    processed++;

                    if (processed % 10 == 0 || processed == total)
                    {
                        Console.WriteLine($"\nProgress: {processed}/{total} files processed");
                    }
                }
                catch (Exception ex)
                {
                    Console.WriteLine($"⚠ ERROR processing {Path.GetFileName(file)}: {ex.Message}");
                    Console.WriteLine($"Exception Type: {ex.GetType().Name}");
                    Console.WriteLine($"Stack trace: {ex.StackTrace}");
                    failed++;
                }
            }

            Console.WriteLine($"\n=== Import Complete ===");
            Console.WriteLine($"Successfully processed: {processed} files");
            Console.WriteLine($"Failed: {failed} files");
        }
    }
}
