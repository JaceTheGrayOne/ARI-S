# __**Unreal Mod Naming**__
**TL;DR**:
Name your mods: `z_ModName_p.pak`  
- `_p` → marks it as a **Patch** (Patch files *always* have priority over base game files).  
- `z_` → wins alphanumerical sorting ties against other patch files that modify the same assets.
-# These are both case-insensitive

## **Unreal File System (UFS)**
- Unreal uses **UFS (Unreal File System)**: a virtualization layer that maps content to virtual mount points.
- Absolute paths like `Content/YourFolder/Blueprints/...` become virtualized paths like `/Game/YourFolder/Blueprints/...` at runtime.
- These virtualized paths are stored in the Asset Registry (`AssetRegistry.bin`) which indexes assets by these virtual paths; UFS performs lookups via this registry to determine which assets to load and in what order.
- This is the reason Console commands like `Summon` expect `/Game/Blueprints/BP_Asset...` , and not the actual absolute path of `Augusta/Content/Blueprints/BP_Asset...`

## **.PAK Prioritization**
When multiple paks provide the same virtual path, UFS resolves conflicts in this order:
1. **Patch Designation**
   - If the filename contains `_p`, it’s give +1000 priority score to its placement in the manifest.
-#          (yes it assigns a literal integer value to priority)

>   "Pak files designated as a patch will get 1000 added to their priority when loaded at startup. This ensures patched files receive a higher sort order, in effect overriding the same files from other paks. Patch paks can be identified by the '\_P' suffix in their file name."
      [Unreal Engine Primer: Loading Paks At Runtime](https://dev.epicgames.com/community/learning/knowledge-base/D7nL/unreal-engine-primer-loading-content-and-pak-files-at-runtime)

2. **Numeric Sort**
   - Paks are sorted by the first numeric value found in the filename.

3. **Mount Order**
   - The order paks are mounted into the manifest further breaks ties.
-#          (This behavior can be custom and specific to the engine version / game)

4. **Alphanumeric Sort**
   - Remaining ties are resolved alphabetically; later items overwrite earlier ones.  
   - Prefixing with `z_` pushes paks later in alpha order, allowing them to win over other patch paks that modify the same files.
   - This sorting is Lexicongraphical which is a fancy way of saying it's sorted exactly how the dictionary is, letter by letter comparison. That means more z's = further down the list.

### Conflict Resolution Example
```
Game.pak                  # Base 
Overhaul_p.pak            # Patch → prioritized over Base
z_PlayerMod_p.pak         # Patch + 'z' → prioritized over Base and sorted later alphabetically over Overhaul
```
Outcome:
`z_PlayerMod_p.pak`’s version of `/Game/.../BP_SurvivalPlayerCharacter` overwrites the one from `Overhaul_p.pak` but the remaining modifications from `Overhaul_p.pak` remain intact and coninue to override the base game assets.