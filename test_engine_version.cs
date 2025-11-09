using System;
using UAssetAPI.UnrealTypes;

class Test {
    static void Main() {
        Console.WriteLine($"VER_UE5_4 = {(int)EngineVersion.VER_UE5_4}");
        Console.WriteLine($"VER_UE5_3 = {(int)EngineVersion.VER_UE5_3}");
        Console.WriteLine($"VER_UE5_2 = {(int)EngineVersion.VER_UE5_2}");
        Console.WriteLine($"VER_UE5_1 = {(int)EngineVersion.VER_UE5_1}");
        Console.WriteLine($"VER_UE5_0 = {(int)EngineVersion.VER_UE5_0}");
        Console.WriteLine($"VER_UE4_27 = {(int)EngineVersion.VER_UE4_27}");
    }
}
