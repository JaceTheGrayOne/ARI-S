// CSS
import './style.css';

// Wails bindings
import { App } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/app";
import { RetocService, RetocOperation, RetocResult } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/retoc";
import { UAssetService, UAssetResult } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/uasset";
import { InjectorService } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/injector";
import { UWPDumperService } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/uwpdumper";
import { Events } from "@wailsio/runtime";

// App state
let currentPane = 'home';
let settingsCache = {
    reduce_motion: false,
    logs_clear_on_launch: true,
    logs_max_lines: 0,
    dir_mods: '',
    dir_exports: '',
    dir_imports: '',
    usmap_path: '',
    remember_paths: true,
    proc_hide_system: false,
    proc_sort: 'name'
};

// Cache of running processes for injector
let processesCache = [];

// Current Retoc operation id
let currentOperationID = null;

let userSetOutputPaths = {
    pak_output_dir: false,
    extract_output_dir: false
};

let firstLoad = {
    retoc: true,
    uasset: true
};

// Initialize app
document.addEventListener('DOMContentLoaded', async function() {
    try {
        const pref = await App.GetPreference('logs_clear_on_launch');
        const shouldClear = (pref === null || pref === undefined) ? true : (String(pref) !== 'false');
        if (shouldClear) {
            const outputs = ['retoc-output','uasset-output','injector-output','uwpdumper-output'];
            outputs.forEach(id => {
                const el = document.getElementById(id);
                if (el) {
                    el.innerHTML = '';
                    const idle = document.createElement('div');
                    idle.className = 'output-line';
                    idle.textContent = 'Idle.';
                    el.appendChild(idle);
                }
            });
        }
    } catch(e) {
        console.warn('Failed to init consoles on load:', e);
    }

    // Initialize UI
    initializeNavigation();
    initializeRetocManager();
    initializeUAssetManager();
    initializeInjectorManager();
    initializeUWPDumperManager();
    initializeSettings();
    initializeMarkdownGuide();
    await loadSettings();
    initializeDragAndDrop();

    document.querySelectorAll('svg.icon').forEach(svg => {
        svg.setAttribute('aria-hidden', 'true');
        svg.setAttribute('focusable', 'false');
    });

    Events.On("retoc:started", (data) => {
        currentOperationID = data.operation_id;
        console.log(`Operation started with ID: ${currentOperationID}`);
    });

    Events.On("retoc:output", (data) => {
        if (data.operation_id === currentOperationID) {
            showOutput('retoc-output', data.line, 'info');
        }
    });

    // Load config
    try {
        const initialConfig = await App.GetInitialConfig();
        await populateFieldsFromConfig(initialConfig);
        console.log('Initial configuration loaded and applied');
    } catch (error) {
        console.error('Failed to load initial configuration:', error);
    }

    switchPane('home');
});

async function populateFieldsFromConfig(config) {

    if (config.input_mod_folder) {
        document.getElementById('input-mod-folder').value = config.input_mod_folder;
    }
    if (config.pak_output_dir) {
        document.getElementById('pak-output-dir').value = config.pak_output_dir;

        if (await App.ValidateDirectory(config.pak_output_dir)) {
            userSetOutputPaths.pak_output_dir = true;
        }
    }
    if (config.game_paks_folder_unpak) {
        document.getElementById('game-paks-folder-unpak').value = config.game_paks_folder_unpak;
    }
    if (config.extract_output_dir) {
        document.getElementById('extract-output-dir').value = config.extract_output_dir;

        if (await App.ValidateDirectory(config.extract_output_dir)) {
            userSetOutputPaths.extract_output_dir = true;
        }
    }


    if (config.export_folder) {
        document.getElementById('export-folder').value = config.export_folder;
        updateExportFileCount();
    }
    if (config.import_folder) {
        document.getElementById('import-folder').value = config.import_folder;
        updateImportFileCount();
    }
    if (config.uasset_mappings_path) {
        document.getElementById('uasset-mappings-path').value = config.uasset_mappings_path;
    }


    if (config.dll_path) {
        document.getElementById('dll-path').value = config.dll_path;
    }


    if (config.ue_version) {
        document.getElementById('ue-version').value = config.ue_version;
    }
}


function initializeNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    const navArray = Array.from(navItems);
    const tiles = document.querySelectorAll('.tile');
    

    navItems.forEach(item => {
        item.setAttribute('role', 'button');
        item.setAttribute('tabindex', '0');
        item.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                item.click();
                return;
            }
            if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
                e.preventDefault();
                const i = navArray.indexOf(item);
                if (i === -1) return;
                const nextIndex = e.key === 'ArrowDown' ? Math.min(i + 1, navArray.length - 1) : Math.max(i - 1, 0);
                navArray[nextIndex]?.focus();
            }
        });

        item.addEventListener('click', () => {
            const pane = item.dataset.pane;
            switchPane(pane);
        });
    });
    

    tiles.forEach(tile => {
        tile.setAttribute('role', 'button');
        tile.setAttribute('tabindex', '0');
        tile.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' || e.key === ' ') {
                e.preventDefault();
                tile.click();
            }
        });

        tile.addEventListener('click', () => {
            const pane = tile.dataset.pane;
            switchPane(pane);
        });
    });
}

function animatePaneElements(pane) {
  const fadeEls = pane.querySelectorAll('.reveal-fade, .reveal-scale');
  fadeEls.forEach((el, i) => {
    el.style.animationDelay = `${i * 0.10}s`;
    el.style.animation = 'none';
    void el.offsetWidth;
    el.style.animation = '';
  });
}

function switchPane(paneName) {
  document.querySelectorAll('.nav-item').forEach(item => {
    item.classList.remove('active');
  });
  document.querySelector(`[data-pane="${paneName}"]`)?.classList.add('active');

  document.querySelectorAll('.pane').forEach(pane => {
    pane.classList.remove('active');
  });
  const next = document.getElementById(`${paneName}-pane`);
  if (!next) return;
  next.classList.add('active');

  currentPane = paneName;

  if (paneName === 'retoc') {
    loadRetocSettings();
  } else if (paneName === 'uasset') {
    loadUAssetSettings();
  } else if (paneName === 'injector') {
    loadProcessList();
  } else if (paneName === 'uwpdumper') {
    loadUWPDumperInfo();
  }

  animatePaneElements(next);
}


// Retoc Manager
function initializeRetocManager() {
    // Browse buttons
    document.getElementById('browse-input-mod').addEventListener('click', () => browseFolder('input-mod-folder'));
    document.getElementById('browse-pak-output').addEventListener('click', () => browseFolder('pak-output-dir'));
    document.getElementById('browse-game-paks-unpak').addEventListener('click', () => browseFolder('game-paks-folder-unpak'));
    document.getElementById('browse-extract-output').addEventListener('click', () => browseFolder('extract-output-dir'));

    document.getElementById('pak-output-dir').addEventListener('input', () => {
        userSetOutputPaths.pak_output_dir = true;
    });
    document.getElementById('extract-output-dir').addEventListener('input', () => {
        userSetOutputPaths.extract_output_dir = true;
    });

    document.getElementById('run-retoc').addEventListener('click', runRetocToZen);
    document.getElementById('run-unpak').addEventListener('click', runRetocUnpak);
    document.getElementById('cancel-unpak').addEventListener('click', cancelRetocOperation);
}

async function loadRetocSettings() {

    if (firstLoad.retoc) {
        const inputModFolder = document.getElementById('input-mod-folder').value;
        const pakOutputDir = document.getElementById('pak-output-dir').value;
        const gamePaksFolderUnpak = document.getElementById('game-paks-folder-unpak').value;
        const extractOutputDir = document.getElementById('extract-output-dir').value;

        if (inputModFolder) {
            showOutput('retoc-output', `Modified UAsset Input Path Loaded: "${inputModFolder}"`, 'info');
        }
        if (pakOutputDir) {
            showOutput('retoc-output', `Mod Output Path Loaded: "${pakOutputDir}"`, 'info');
        }
        if (gamePaksFolderUnpak) {
            showOutput('retoc-output', `Game Paks Path Loaded: "${gamePaksFolderUnpak}"`, 'info');
        }
        if (extractOutputDir) {
            showOutput('retoc-output', `Extracted Asset Path Loaded: "${extractOutputDir}"`, 'info');
        }
        firstLoad.retoc = false;
    }

    try {
        if (!document.getElementById('pak-output-dir').value && settingsCache.dir_mods) {
            document.getElementById('pak-output-dir').value = settingsCache.dir_mods;
            userSetOutputPaths.pak_output_dir = true;
        }
        if (!document.getElementById('input-mod-folder').value && settingsCache.dir_mods) {
            document.getElementById('input-mod-folder').value = settingsCache.dir_mods;
        }
        if (!document.getElementById('extract-output-dir').value && settingsCache.dir_imports) {
            document.getElementById('extract-output-dir').value = settingsCache.dir_imports;
            userSetOutputPaths.extract_output_dir = true;
        }
    } catch {}
}

async function runRetocToZen() {
    const inputPath = document.getElementById('input-mod-folder').value;
    const outputPath = document.getElementById('pak-output-dir').value;
    const ueVersion = document.getElementById('ue-version').value;
    const outputBaseName = document.getElementById('mod-name').value.trim();
    const orderSuffix = document.getElementById('load-order').value.trim();

    if (!inputPath || !outputPath) {
        showOutput('retoc-output', 'Error: Please specify input and output paths.', 'error');
        return;
    }

    // Validate that output path set
    if (!userSetOutputPaths.pak_output_dir) {
        showOutput('retoc-output', 'Error: Please explicitly specify an output directory using the browse button or by typing a valid path.', 'error');
        return;
    }

    // Validate that output directory exists
    const isValidOutput = await App.ValidateDirectory(outputPath);
    if (!isValidOutput) {
        showOutput('retoc-output', 'Error: Output directory does not exist or is not accessible. Please select a valid directory.', 'error');
        return;
    }

    // Validate serialization number
    if (orderSuffix && !/^\d+$/.test(orderSuffix)) {
        showOutput('retoc-output', 'Error: Serialization number must contain only digits 0-9.', 'error');
        return;
    }

    // Validate mod name
    // Invalid characters: < > : " / \ | ? *
    if (outputBaseName && /[<>:"/\\|?*]/.test(outputBaseName)) {
        showOutput('retoc-output', 'Error: Mod name contains invalid characters. Cannot use: < > : " / \\ | ? *', 'error');
        return;
    }

    // Save paths
    App.SetLastUsedPath('input_mod_folder', inputPath);
    App.SetLastUsedPath('pak_output_dir', outputPath);
    App.SetPreference('ue_version', ueVersion);

    showOutput('retoc-output', 'Starting Retoc to-zen operation...', 'info');

    // Build operation
    const operation = new RetocOperation({
        command: 'to-zen',
        input_path: inputPath,
        output_path: outputPath,
        ue_version: ueVersion,
        options: []
    });

    // Add mod name and serialization
    if (outputBaseName) {
        operation.options.push('--mod-name', outputBaseName);
    }

    if (orderSuffix) {
        operation.options.push('--serialization', orderSuffix);
    }
    
    // Execute operation
    RetocService.RunRetoc(operation)
        .then(result => {
            if (result.success) {
                showOutput('retoc-output', `Success: ${result.message}`, 'success');
                showOutput('retoc-output', result.output, 'info');
            } else {
                showOutput('retoc-output', `Error: ${result.error}`, 'error');
                showOutput('retoc-output', result.output, 'error');
            }
        })
        .catch(error => {
            showOutput('retoc-output', `Error: ${error.message}`, 'error');
        });
}

async function runRetocUnpak() {
    const gamePaksFolder = document.getElementById('game-paks-folder-unpak').value;
    const extractOutputDir = document.getElementById('extract-output-dir').value;

    if (!gamePaksFolder || !extractOutputDir) {
        showOutput('retoc-output', 'Error: Please specify game paks folder and extract output directory.', 'error');
        return;
    }

    // Validate that output path was set
    if (!userSetOutputPaths.extract_output_dir) {
        showOutput('retoc-output', 'Error: Please explicitly specify an output directory using the browse button or by typing a valid path.', 'error');
        return;
    }

    // Validate that output directory exists
    const isValidOutput = await App.ValidateDirectory(extractOutputDir);
    if (!isValidOutput) {
        showOutput('retoc-output', 'Error: Output directory does not exist or is not accessible. Please select a valid directory.', 'error');
        return;
    }

    // Save paths to App
    App.SetLastUsedPath('game_paks_folder_unpak', gamePaksFolder);
    App.SetLastUsedPath('extract_output_dir', extractOutputDir);

    document.getElementById('run-unpak').style.display = 'none';
    document.getElementById('cancel-unpak').style.display = 'inline-block';

    showOutput('retoc-output', 'Starting Retoc extraction operation (to-legacy)...', 'info');
    showOutput('retoc-output', 'This is working, the blank cmd shell is normal.', 'info');

    // Build operation
    const operation = new RetocOperation({
        command: 'to-legacy',
        input_path: gamePaksFolder,
        output_path: extractOutputDir,
        options: []
    });

    // Execute operation
    RetocService.RunRetoc(operation)
        .then(result => {
            document.getElementById('run-unpak').style.display = 'inline-block';
            document.getElementById('cancel-unpak').style.display = 'none';

            if (result.success) {
                showOutput('retoc-output', `Success: ${result.message}`, 'success');
                if (result.output) {
                    showOutput('retoc-output', result.output, 'info');
                }
            } else {
                if (result.error === 'cancelled') {
                    showOutput('retoc-output', 'Operation cancelled by user', 'warning');
                } else {
                    showOutput('retoc-output', `Error: ${result.error}`, 'error');
                    if (result.output) {
                        showOutput('retoc-output', result.output, 'error');
                    }
                }
            }
        })
        .catch(error => {
            document.getElementById('run-unpak').style.display = 'inline-block';
            document.getElementById('cancel-unpak').style.display = 'none';
            currentOperationID = null;

            showOutput('retoc-output', `Error: ${error.message}`, 'error');
        });
}

function cancelRetocOperation() {
    showOutput('retoc-output', 'Cancelling operation...', 'warning');

    // Cancel to-legacy extraction operation
    RetocService.CancelOperationByCommand('to-legacy')
        .then(() => {
            showOutput('retoc-output', 'Cancellation requested', 'info');
        })
        .catch(error => {
            showOutput('retoc-output', `Failed to cancel: ${error.message}`, 'error');
        });
}


// UAsset Manager
function initializeUAssetManager() {
    // Browse buttons
    document.getElementById('browse-export-folder').addEventListener('click', () => browseFolder('export-folder'));
    document.getElementById('browse-import-folder').addEventListener('click', () => browseFolder('import-folder'));
    document.getElementById('browse-uasset-mappings').addEventListener('click', () => browseFile('uasset-mappings-path'));

    // Action buttons
    document.getElementById('export-uassets').addEventListener('click', exportUAssets);
    document.getElementById('import-uassets').addEventListener('click', importUAssets);

    // File count updates
    document.getElementById('export-folder').addEventListener('input', updateExportFileCount);
    document.getElementById('import-folder').addEventListener('input', updateImportFileCount);
}

async function loadUAssetSettings() {

    if (firstLoad.uasset) {
        const mappingsPath = document.getElementById('uasset-mappings-path').value;
        const exportFolder = document.getElementById('export-folder').value;
        const importFolder = document.getElementById('import-folder').value;

        if (mappingsPath) {
            showOutput('uasset-output', `Mapping File Loaded: "${mappingsPath}"`, 'info');
        }
        if (exportFolder) {
            showOutput('uasset-output', `Export Path Loaded: "${exportFolder}"`, 'info');
        }
        if (importFolder) {
            showOutput('uasset-output', `Import Path Loaded: "${importFolder}"`, 'info');
        }
        firstLoad.uasset = false;
    }

    try {
        if (!document.getElementById('export-folder').value && settingsCache.dir_exports) {
            document.getElementById('export-folder').value = settingsCache.dir_exports;
            updateExportFileCount();
        }
        if (!document.getElementById('import-folder').value && settingsCache.dir_imports) {
            document.getElementById('import-folder').value = settingsCache.dir_imports;
            updateImportFileCount();
        }
        if (!document.getElementById('uasset-mappings-path').value && settingsCache.usmap_path) {
            document.getElementById('uasset-mappings-path').value = settingsCache.usmap_path;
        }
    } catch {}
}

function exportUAssets() {
    const folderPath = document.getElementById('export-folder').value;
    const mappingsPath = document.getElementById('uasset-mappings-path').value;

    if (!folderPath) {
        showOutput('uasset-output', 'Error: Please select a folder containing .uasset/.uexp files.', 'error');
        return;
    }

    // Save paths to App
    App.SetLastUsedPath('export_folder', folderPath);
    if (mappingsPath) {
        App.SetLastUsedPath('uasset_mappings_path', mappingsPath);
    }

    showOutput('uasset-output', 'Starting UAsset export to JSON...', 'info');
    if (mappingsPath) {
        showOutput('uasset-output', `Using mappings file: ${mappingsPath}`, 'info');
    } else {
        showOutput('uasset-output', 'No mappings file provided - export may be incomplete', 'warning');
    }

    UAssetService.ExportUAssets(folderPath, mappingsPath)
        .then(result => {
            if (result.success) {
                showOutput('uasset-output', `Success: ${result.message}`, 'success');
                showOutput('uasset-output', `Files processed: ${result.files_processed}`, 'info');
                showOutput('uasset-output', result.output, 'info');
                updateExportFileCount();
            } else {
                showOutput('uasset-output', `Error: ${result.error}`, 'error');
                showOutput('uasset-output', result.output, 'error');
            }
        })
        .catch(error => {
            showOutput('uasset-output', `Error: ${error.message}`, 'error');
        });
}

function importUAssets() {
    const folderPath = document.getElementById('import-folder').value;
    const mappingsPath = document.getElementById('uasset-mappings-path').value;

    if (!folderPath) {
        showOutput('uasset-output', 'Error: Please select a folder containing .json files.', 'error');
        return;
    }

    // Save paths to App
    App.SetLastUsedPath('import_folder', folderPath);
    if (mappingsPath) {
        App.SetLastUsedPath('uasset_mappings_path', mappingsPath);
    }

    showOutput('uasset-output', 'Starting UAsset import from JSON...', 'info');
    if (mappingsPath) {
        showOutput('uasset-output', `Using mappings file: ${mappingsPath}`, 'info');
    } else {
        showOutput('uasset-output', 'Warning: No mappings file provided - import may fail for unversioned properties', 'warning');
    }

    UAssetService.ImportUAssets(folderPath, mappingsPath)
        .then(result => {
            if (result.success) {
                showOutput('uasset-output', `Success: ${result.message}`, 'success');
                showOutput('uasset-output', `Files processed: ${result.files_processed}`, 'info');
                showOutput('uasset-output', result.output, 'info');
                updateImportFileCount();
            } else {
                showOutput('uasset-output', `Error: ${result.error}`, 'error');
                showOutput('uasset-output', result.output, 'error');
            }
        })
        .catch(error => {
            showOutput('uasset-output', `Error: ${error.message}`, 'error');
        });
}

async function updateExportFileCount() {
    const folderPath = document.getElementById('export-folder').value;
    const countElement = document.getElementById('export-file-count').querySelector('.file-count-value');

    if (!folderPath) {
        countElement.textContent = '0 .uasset, 0 .uexp';
        return;
    }

    countElement.textContent = 'Counting files...';

    try {
        const [uassetCount, uexpCount] = await UAssetService.CountUAssetFiles(folderPath);
        countElement.textContent = `${uassetCount} .uasset, ${uexpCount} .uexp`;
    } catch (error) {
        countElement.textContent = 'Error counting files';
        console.error('File counting failed:', error);
    }
}

async function updateImportFileCount() {
    const folderPath = document.getElementById('import-folder').value;
    const countElement = document.getElementById('import-file-count').querySelector('.file-count-value');

    if (!folderPath) {
        countElement.textContent = '0 .json';
        return;
    }

    countElement.textContent = 'Counting files...';

    try {
        const jsonCount = await UAssetService.CountJSONFiles(folderPath);
        countElement.textContent = `${jsonCount} .json`;
    } catch (error) {
        countElement.textContent = 'Error counting files';
        console.error('File counting failed:', error);
    }
}

// DLL Injector Manager
function initializeInjectorManager() {
    // Browse button
    document.getElementById('browse-dll').addEventListener('click', () => browseDLL('dll-path'));

    // Refresh button
    document.getElementById('refresh-processes').addEventListener('click', loadProcessList);

    // Process selector
    document.getElementById('process-selector').addEventListener('change', (e) => {
        const selector = e.target;
        const processInfo = document.getElementById('process-info');
        const selectedPidElement = document.getElementById('selected-pid');

        if (selector.value) {
            selectedPidElement.textContent = selector.value;
            processInfo.style.display = 'block';
        } else {
            processInfo.style.display = 'none';
        }
    });

    // Process filter
    document.getElementById('process-filter').addEventListener('input', () => {
        applyProcessFilter();
    });

    // Inject button
    document.getElementById('inject-dll').addEventListener('click', injectDLL);

    // Dump SDK button
    document.getElementById('dump-sdk').addEventListener('click', dumpSDK);

    // Dump Mappings button
    document.getElementById('dump-mappings').addEventListener('click', dumpMappings);
}

async function browseDLL(inputId) {
    try {
        const title = 'Select DLL file to inject';
        const filter = "DLL Files\x00*.dll\x00All Files\x00*.*\x00\x00";
        const key = inputId.replace(/-/g, '_');

        const result = await App.BrowseFile(title, filter, key);
        if (result) {
            document.getElementById(inputId).value = result;
            if (settingsCache.remember_paths) {
                await App.SetLastUsedPath(key, result);
            }
            showOutput('injector-output', `DLL selected: ${result}`, 'info');
        }
    } catch (error) {
        console.error('Error browsing DLL file:', error);
        showOutput('injector-output', `Error selecting DLL: ${error.message}`, 'error');
    }
}

async function loadProcessList() {
    showOutput('injector-output', 'Loading running processes...', 'info');

    try {
        let processes = await InjectorService.GetRunningProcesses();
        const selector = document.getElementById('process-selector');
        const filterInput = document.getElementById('process-filter');

        // Clear filter
        if (filterInput) {
            filterInput.value = '';
        }

        // Clear existing options
        selector.innerHTML = '<option value="">-- Select a process --</option>';

        if (settingsCache.proc_hide_system) {
            const name = (s)=>String(s||'').toLowerCase();
            const blacklistExact = new Set(['system','idle','registry','smss.exe','csrss.exe','wininit.exe','services.exe','lsass.exe','winlogon.exe','fontdrvhost.exe','conhost.exe','spoolsv.exe','dwm.exe','sihost.exe','ctfmon.exe','securityhealthservice.exe']);
            const patterns = [/^svchost\.exe$/i,/host\.exe$/i,/broker/i,/search/i,/experiencehost/i,/runtimebroker/i,/startmenuexperiencehost/i,/shellexperiencehost/i,/textinputhost/i,/systemsettings/i,/snippingtool/i];
            processes = processes.filter(p => {
                const n = name(p.name);
                if (blacklistExact.has(n)) return false;
                return !patterns.some(rx => rx.test(n));
            });
        }
        if (settingsCache.proc_sort === 'pid') {
            processes.sort((a, b) => (a.pid - b.pid));
        } else {
            processes.sort((a, b) => String(a.name).localeCompare(String(b.name)));
        }

        processesCache = processes;
        selector.innerHTML = '<option value="">-- Select a process --</option>';
        processesCache.forEach(p => {
            const opt = document.createElement('option');
            opt.value = String(p.pid);
            opt.textContent = `${String(p.name || 'Unknown')} (PID: ${p.pid})`;
            selector.appendChild(opt);
        });

        showOutput('injector-output', `Loaded ${processes.length} running processes`, 'success');
    } catch (error) {
        console.error('Error loading processes:', error);
        showOutput('injector-output', `Error loading processes: ${error.message}`, 'error');
    }
}



function applyProcessFilter() {
    const selector = document.getElementById('process-selector');
    const needle = (document.getElementById('process-filter')?.value || '').toLowerCase();
    selector.innerHTML = '<option value="">-- Select a process --</option>';
    processesCache.forEach(p => {
        const label = `${String(p.name || 'Unknown')} (PID: ${p.pid})`;
        if (needle && !label.toLowerCase().includes(needle)) return;
        const opt = document.createElement('option');
        opt.value = String(p.pid);
        opt.textContent = label;
        selector.appendChild(opt);
    });
}
async function injectDLL() {
    const dllPath = document.getElementById('dll-path').value;
    const selector = document.getElementById('process-selector');
    const selectedPID = selector.value;

    // Validation
    if (!dllPath) {
        showOutput('injector-output', 'Error: Please select a DLL file to inject.', 'error');
        return;
    }

    if (!selectedPID) {
        showOutput('injector-output', 'Error: Please select a target process.', 'error');
        return;
    }

    // Get process name for display
    const selectedOption = selector.options[selector.selectedIndex];
    const processName = selectedOption.textContent;

    showOutput('injector-output', `Attempting to inject ${dllPath}`, 'info');
    showOutput('injector-output', `Target: ${processName}`, 'info');
    showOutput('injector-output', '---', 'info');

    // Disable inject button during operation
    const injectButton = document.getElementById('inject-dll');
    injectButton.disabled = true;
    injectButton.textContent = 'Injecting...';

    try {
        // Call the injection service
        const result = await InjectorService.InjectDLL(parseInt(selectedPID), dllPath);

        if (result.success) {
            showOutput('injector-output', result.message, 'success');
            showOutput('injector-output', result.output, 'info');
            showOutput('injector-output', `Duration: ${result.duration}`, 'info');
        } else {
            if (result.error === 'NEEDS_ELEVATION') {
                showOutput('injector-output', 'âš ï¸ ADMINISTRATOR PRIVILEGES REQUIRED', 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', result.output, 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', 'Your DLL path has been saved.', 'success');
                showOutput('injector-output', 'After restart, simply click "Inject DLL" again.', 'info');
            } else {
                showOutput('injector-output', `Injection failed: ${result.message}`, 'error');
                if (result.error) {
                    showOutput('injector-output', `Error details: ${result.error}`, 'error');
                }
                if (result.output) {
                    showOutput('injector-output', result.output, 'error');
                }
            }
        }
    } catch (error) {
        showOutput('injector-output', `Injection error: ${error.message}`, 'error');
        console.error('Injection failed:', error);
    } finally {
        injectButton.disabled = false;
        injectButton.textContent = 'Inject DLL';
    }
}

async function dumpSDK() {
    const selector = document.getElementById('process-selector');
    const selectedPID = selector.value;

    // Validation
    if (!selectedPID) {
        showOutput('injector-output', 'Error: Please select a target process first.', 'error');
        return;
    }

    // Get process name for display
    const selectedOption = selector.options[selector.selectedIndex];
    const processName = selectedOption.textContent;

    // Get Dumper7.dll
    let dumper7Path;
    try {
        dumper7Path = await App.GetDumper7Path();
        if (!dumper7Path) {
            showOutput('injector-output', 'Error: Could not locate Dumper7.dll in dependencies directory.', 'error');
            return;
        }
    } catch (error) {
        showOutput('injector-output', `Error getting Dumper7 path: ${error.message}`, 'error');
        return;
    }

    showOutput('injector-output', `Attempting to inject Dumper7 SDK dumper`, 'info');
    showOutput('injector-output', `DLL: ${dumper7Path}`, 'info');
    showOutput('injector-output', `Target: ${processName}`, 'info');
    showOutput('injector-output', '---', 'info');

    // Disable dump button during operation
    const dumpButton = document.getElementById('dump-sdk');
    dumpButton.disabled = true;
    dumpButton.textContent = 'Dumping...';

    try {
        // Call the injection service with Dumper7.dll
        const result = await InjectorService.InjectDLL(parseInt(selectedPID), dumper7Path);

        if (result.success) {
            showOutput('injector-output', result.message, 'success');
            showOutput('injector-output', result.output, 'info');
            showOutput('injector-output', `Duration: ${result.duration}`, 'info');
            showOutput('injector-output', '---', 'info');
            showOutput('injector-output', 'SDK dump initiated. Check the game directory for output files.', 'success');
        } else {
            if (result.error === 'NEEDS_ELEVATION') {
                showOutput('injector-output', 'âš ï¸ ADMINISTRATOR PRIVILEGES REQUIRED', 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', result.output, 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', 'Please restart the application as Administrator and try again.', 'info');
            } else {
                showOutput('injector-output', `SDK dump failed: ${result.message}`, 'error');
                if (result.error) {
                    showOutput('injector-output', `Error details: ${result.error}`, 'error');
                }
                if (result.output) {
                    showOutput('injector-output', result.output, 'error');
                }
            }
        }
    } catch (error) {
        showOutput('injector-output', `SDK dump error: ${error.message}`, 'error');
        console.error('SDK dump failed:', error);
    } finally {
        dumpButton.disabled = false;
        dumpButton.textContent = 'Dump SDK';
    }
}

async function dumpMappings() {
    const selector = document.getElementById('process-selector');
    const selectedPID = selector.value;

    // Validation
    if (!selectedPID) {
        showOutput('injector-output', 'Error: Please select a target process first.', 'error');
        return;
    }

    // Get process name for display
    const selectedOption = selector.options[selector.selectedIndex];
    const processName = selectedOption.textContent;

    // Get UnrealMappingsDumper.dll
    let mappingsDumperPath;
    try {
        mappingsDumperPath = await App.GetUnrealMappingsDumperPath();
        if (!mappingsDumperPath) {
            showOutput('injector-output', 'Error: Could not locate UnrealMappingsDumper.dll in dependencies directory.', 'error');
            return;
        }
    } catch (error) {
        showOutput('injector-output', `Error getting UnrealMappingsDumper path: ${error.message}`, 'error');
        return;
    }

    showOutput('injector-output', `Attempting to inject UnrealMappingsDumper`, 'info');
    showOutput('injector-output', `DLL: ${mappingsDumperPath}`, 'info');
    showOutput('injector-output', `Target: ${processName}`, 'info');
    showOutput('injector-output', '---', 'info');

    // Disable dump button during operation
    const dumpButton = document.getElementById('dump-mappings');
    dumpButton.disabled = true;
    dumpButton.textContent = 'Dumping...';

    try {
        // Call the injection service with UnrealMappingsDumper.dll
        const result = await InjectorService.InjectDLL(parseInt(selectedPID), mappingsDumperPath);

        if (result.success) {
            showOutput('injector-output', result.message, 'success');
            showOutput('injector-output', result.output, 'info');
            showOutput('injector-output', `Duration: ${result.duration}`, 'info');
            showOutput('injector-output', '---', 'info');
            showOutput('injector-output', 'Mappings dump initiated. Check the game directory for .usmap output file.', 'success');
        } else {
            if (result.error === 'NEEDS_ELEVATION') {
                showOutput('injector-output', 'âš ï¸ ADMINISTRATOR PRIVILEGES REQUIRED', 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', result.output, 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', 'Please restart the application as Administrator and try again.', 'info');
            } else {
                showOutput('injector-output', `Mappings dump failed: ${result.message}`, 'error');
                if (result.error) {
                    showOutput('injector-output', `Error details: ${result.error}`, 'error');
                }
                if (result.output) {
                    showOutput('injector-output', result.output, 'error');
                }
            }
        }
    } catch (error) {
        showOutput('injector-output', `Mappings dump error: ${error.message}`, 'error');
        console.error('Mappings dump failed:', error);
    } finally {
        dumpButton.disabled = false;
        dumpButton.textContent = 'Dump Mappings';
    }
}

// UWPDumper Manager
function initializeUWPDumperManager() {
    // Launch button
    document.getElementById('launch-uwpdumper').addEventListener('click', launchUWPDumper);
}

async function loadUWPDumperInfo() {
    showOutput('uwpdumper-output', 'Loading UWPDumper tool information...', 'info');

    try {
        // Get dumper info from backend
        const info = await UWPDumperService.GetDumperInfo();

        // Update launch button
        const launchButton = document.getElementById('launch-uwpdumper');

        if (info.ready) {
            launchButton.disabled = false;
            showOutput('uwpdumper-output', 'UWPDumper tool is ready to use', 'success');
            showOutput('uwpdumper-output', `Tool location: ${info.dumper_path}`, 'info');
        } else {
            launchButton.disabled = true;
            showOutput('uwpdumper-output', 'UWPDumper tool not found', 'error');
            showOutput('uwpdumper-output', 'Please download UWPDumper binaries and place them in:', 'info');
            showOutput('uwpdumper-output', `  ${info.dumper_path.replace('UWPInjector.exe', '')}`, 'info');
            showOutput('uwpdumper-output', 'See dependencies/uwpdumper/README.md for instructions', 'info');
        }
    } catch (error) {
        console.error('Error loading UWPDumper info:', error);
        showOutput('uwpdumper-output', `Error loading tool info: ${error.message}`, 'error');
        const launchButton = document.getElementById('launch-uwpdumper');
        if (launchButton) {
            launchButton.disabled = true;
        }
    }
}

async function launchUWPDumper() {
    showOutput('uwpdumper-output', 'Launching UWPDumper...', 'info');
    showOutput('uwpdumper-output', '---', 'info');

    // Disable launch button during operation
    const launchButton = document.getElementById('launch-uwpdumper');
    launchButton.disabled = true;
    launchButton.textContent = 'Launching...';

    try {
        // Call the launch service
        const result = await UWPDumperService.LaunchUWPDumper();

        if (result.success) {
            showOutput('uwpdumper-output', result.message, 'success');
            showOutput('uwpdumper-output', '---', 'info');
            showOutput('uwpdumper-output', result.output, 'info');
            showOutput('uwpdumper-output', '---', 'info');
            showOutput('uwpdumper-output', `Tool launched in: ${result.duration}`, 'info');
        } else {
            showOutput('uwpdumper-output', `Launch failed: ${result.message}`, 'error');
            if (result.error) {
                showOutput('uwpdumper-output', `Error details: ${result.error}`, 'error');
            }
            if (result.output) {
                showOutput('uwpdumper-output', result.output, 'info');
            }
        }
    } catch (error) {
        showOutput('uwpdumper-output', `Launch error: ${error.message}`, 'error');
        console.error('Launch failed:', error);
    } finally {
        // Re-enable launch button
        launchButton.disabled = false;
        launchButton.textContent = 'Launch UWPDumper';
    }
}

// Settings
function initializeSettings() {
  document.getElementById('save-settings').addEventListener('click', saveSettings);
  const exp = document.getElementById('pref-export');
  const imp = document.getElementById('pref-import');
  const file = document.getElementById('pref-import-file');
  const reset = document.getElementById('pref-reset');
  if (exp) exp.addEventListener('click', exportSettings);
  if (imp && file) imp.addEventListener('click', () => file.click());
  if (file) file.addEventListener('change', importSettingsFromFile);
  if (reset) reset.addEventListener('click', async () => { await resetDefaults(); await loadSettings(); showNotification('Settings reset to defaults', 'success'); });

  // Browse buttons
  const pairs = [
    ['browse-pref-dir-mods', 'pref-dir-mods', 'folder'],
    ['browse-pref-dir-exports', 'pref-dir-exports', 'folder'],
    ['browse-pref-dir-imports', 'pref-dir-imports', 'folder'],
    ['browse-pref-usmap-path', 'pref-usmap-path', 'file']
  ];
  pairs.forEach(([btnId, inputId, kind]) => {
    const btn = document.getElementById(btnId);
    if (!btn) return;
    btn.addEventListener('click', async () => {
      if (kind === 'folder') {
        await browseFolder(inputId);
      } else {
        await browseFile(inputId);
      }
    });
  });
}

async function loadSettings() {
  const entries = [
    ['reduce_motion','false'],
    ['logs_clear_on_launch','true'],
    ['logs_max_lines','0'],
    ['dir_mods',''],
    ['dir_exports',''],
    ['dir_imports',''],
    ['usmap_path',''],
    ['remember_paths','true'],
    ['proc_hide_system','false'],
    ['proc_sort','name']
  ];
  for (const [key, def] of entries) {
    try { const v = await App.GetPreference(key); settingsCache[key] = (v ?? def); } catch {}
  }

  settingsCache.reduce_motion = settingsCache.reduce_motion === 'true' || settingsCache.reduce_motion === true;
  settingsCache.logs_clear_on_launch = settingsCache.logs_clear_on_launch === 'true' || settingsCache.logs_clear_on_launch === true;
  settingsCache.logs_max_lines = parseInt(settingsCache.logs_max_lines || '0', 10) || 0;
  settingsCache.remember_paths = settingsCache.remember_paths === 'true' || settingsCache.remember_paths === true;
  settingsCache.proc_hide_system = settingsCache.proc_hide_system === 'true' || settingsCache.proc_hide_system === true;

  // UI
  const byId = (id) => document.getElementById(id);
  byId('pref-reduce-motion').checked = settingsCache.reduce_motion;
  byId('pref-logs-clear').checked = settingsCache.logs_clear_on_launch;
  byId('pref-logs-max').value = settingsCache.logs_max_lines;
  byId('pref-dir-mods').value = settingsCache.dir_mods;
  byId('pref-dir-exports').value = settingsCache.dir_exports;
  byId('pref-dir-imports').value = settingsCache.dir_imports;
  byId('pref-usmap-path').value = settingsCache.usmap_path;
  byId('pref-remember-paths').checked = settingsCache.remember_paths;
  byId('pref-proc-hide-system').checked = settingsCache.proc_hide_system;
  byId('pref-proc-sort').value = settingsCache.proc_sort;

  document.body.classList.toggle('reduce-motion', settingsCache.reduce_motion);
}

async function saveSettings() {
  const byId = (id) => document.getElementById(id);
  settingsCache.reduce_motion = byId('pref-reduce-motion').checked;
  settingsCache.logs_clear_on_launch = byId('pref-logs-clear').checked;
  settingsCache.logs_max_lines = parseInt(byId('pref-logs-max').value || '0', 10) || 0;
  settingsCache.dir_mods = byId('pref-dir-mods').value.trim();
  settingsCache.dir_exports = byId('pref-dir-exports').value.trim();
  settingsCache.dir_imports = byId('pref-dir-imports').value.trim();
  settingsCache.usmap_path = byId('pref-usmap-path').value.trim();
  settingsCache.remember_paths = byId('pref-remember-paths').checked;
  settingsCache.proc_hide_system = byId('pref-proc-hide-system').checked;
  settingsCache.proc_sort = byId('pref-proc-sort').value;

  for (const [k, v] of Object.entries(settingsCache)) {
    await App.SetPreference(k, String(v));
  }
  document.body.classList.toggle('reduce-motion', settingsCache.reduce_motion);
  showNotification('Settings saved successfully!', 'success');
}

async function resetDefaults() {
  const defaults = {
    reduce_motion: false,
    logs_clear_on_launch: true,
    logs_max_lines: 0,
    dir_mods: '',
    dir_exports: '',
    dir_imports: '',
    usmap_path: '',
    remember_paths: true,
    proc_hide_system: false,
    proc_sort: 'name'
  };
  for (const [k, v] of Object.entries(defaults)) {
    await App.SetPreference(k, String(v));
  }
}

function exportSettings() {
  const data = settingsCache;
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a'); a.href = url; a.download = 'aris-settings-backup.json';
  document.body.appendChild(a); a.click(); a.remove(); URL.revokeObjectURL(url);
}

async function importSettingsFromFile(e) {
  const file = e.target.files?.[0];
  if (!file) return;
  try {
    const text = await file.text();
    const obj = JSON.parse(text);
    for (const [k, v] of Object.entries(obj)) {
      if (k in settingsCache) await App.SetPreference(k, String(v));
    }
    await loadSettings();
    showNotification('Settings restored from backup', 'success');
  } catch (err) {
    console.error('Import failed:', err);
    showNotification('Failed to restore settings: invalid file', 'error');
  } finally {
    e.target.value = '';
  }
}

// Utility
async function browseFolder(inputId) {
    try {
        const title = `Select folder for ${inputId.replace('-', ' ')}`;
        const key = inputId.replace(/-/g, '_');

        const result = await App.BrowseFolder(title, key);
        if (result) {
            document.getElementById(inputId).value = result;

            if (key === 'pak_output_dir' || key === 'extract_output_dir') {
                userSetOutputPaths[key] = true;
            }

            if (inputId === 'export-folder') {
                updateExportFileCount();
            } else if (inputId === 'import-folder') {
                updateImportFileCount();
            }

            if (settingsCache.remember_paths) {
                await App.SetLastUsedPath(key, result);
            }
        }
    } catch (error) {
        console.error('Error browsing folder:', error);
    }
}

async function browseFile(inputId) {
    try {
        const title = `Select file for ${inputId.replace('-', ' ')}`;
        const filter = "USMAP Files\x00*.usmap\x00All Files\x00*.*\x00\x00";

        const key = inputId.replace(/-/g, '_');

        const result = await App.BrowseFile(title, filter, key);
        if (result) {
            document.getElementById(inputId).value = result;
            if (settingsCache.remember_paths) {
                await App.SetLastUsedPath(key, result);
            }
        }
    } catch (error) {
        console.error('Error browsing file:', error);
    }
}

function showOutput(containerId, message, type = 'info') {
    const container = document.getElementById(containerId);
    const line = document.createElement('div');
    line.className = `output-line ${type}`;
    line.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
    container.appendChild(line);
    container.scrollTop = container.scrollHeight;
    const max = parseInt(settingsCache?.logs_max_lines || 0, 10);
    if (max > 0) {
        while (container.childElementCount > max) {
            container.removeChild(container.firstElementChild);
        }
    }
}

function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 12px 16px;
        border-radius: 8px;
        color: white;
        font-size: 14px;
        font-weight: 500;
        z-index: 10000;
        opacity: 0;
        transform: translateX(100%);
        transition: all 0.3s ease;
        max-width: 300px;
        word-wrap: break-word;
    `;
    
    switch (type) {
        case 'success':
            notification.style.background = 'var(--gradient-success)';
            break;
        case 'error':
            notification.style.background = 'var(--gradient-error)';
            break;
        case 'warning':
            notification.style.background = 'var(--gradient-warning)';
            break;
        default:
            notification.style.background = 'var(--gradient-accent)';
    }
    
    document.body.appendChild(notification);
    
    setTimeout(() => {
        notification.style.opacity = '1';
        notification.style.transform = 'translateX(0)';
    }, 10);
    
    setTimeout(() => {
        notification.style.opacity = '0';
        notification.style.transform = 'translateX(100%)';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, 3000);
}

// Drag and Drop
function initializeDragAndDrop() {
    const dropZones = document.querySelectorAll('.form-input[type="text"]');
    
    dropZones.forEach(zone => {
        zone.addEventListener('dragover', (e) => {
            e.preventDefault();
            zone.classList.add('drag-over');
        });
        
        zone.addEventListener('dragleave', () => {
            zone.classList.remove('drag-over');
        });
        
        zone.addEventListener('drop', (e) => {
            e.preventDefault();
            zone.classList.remove('drag-over');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                const file = files[0];
                if (file.type === 'application/octet-stream' || file.name.endsWith('.pak')) {
                    zone.value = file.path || file.name;
                }
            }
        });
    });
}


// ----- Markdown Guide -----
function initializeMarkdownGuide() {
  const openBtn = document.getElementById('open-naming-guide');
  const modal = document.getElementById('md-modal');
  const closeBtn = document.getElementById('md-modal-close');
  const content = document.getElementById('md-modal-content');
  if (!openBtn || !modal || !closeBtn || !content) return;

  const open = async () => {
    try {
      const res = await fetch('UE_Mod_Naming.md', { cache: 'no-store' });
      const md = await res.text();
      let html = '';
      try {
        html = await App.RenderMarkdown(md);
      } catch (err) {
        html = '';
      }
      content.innerHTML = html || renderMarkdownBasic(md);
      modal.style.display = 'flex';
    } catch (err) {
      content.innerHTML = '<p style="color:#ff6b6b">Failed to load guide.</p>';
      modal.style.display = 'flex';
    }
  };
  const close = () => { modal.style.display = 'none'; };

  openBtn.addEventListener('click', open);
  closeBtn.addEventListener('click', close);
  modal.addEventListener('click', (e) => { if (e.target === modal) close(); });
  document.addEventListener('keydown', (e) => { if (e.key === 'Escape') close(); });
}

function renderMarkdownBasic(md) {
  let s = md.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  s = s.replace(/```([\s\S]*?)```/g, function(_, code){ return '<pre><code>' + code + '</code></pre>'; });
  s = s.replace(/^######\s+(.*)$/gm, '<h6>$1</h6>')
       .replace(/^#####\s+(.*)$/gm, '<h5>$1</h5>')
       .replace(/^####\s+(.*)$/gm, '<h4>$1</h4>')
       .replace(/^###\s+(.*)$/gm, '<h3>$1</h3>')
       .replace(/^##\s+(.*)$/gm, '<h2>$1</h2>')
       .replace(/^#\s+(.*)$/gm, '<h1>$1</h1>');
  s = s.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
  s = s.replace(/\*(.*?)\*/g, '<em>$1</em>');
  s = s.replace(/^\-\s+(.*)$/gm, '<li>$1</li>');
  s = s.replace(/(?:<li>.*?<\/li>\r?\n?)+/g, function(m){ return '<ul>' + m + '</ul>'; });
  var parts = s.split(/\n\n+/).map(function(block){
    if (/^<h[1-6]>/.test(block) || /^<pre>/.test(block) || /^<ul>/.test(block)) return block;
    return '<p>' + block.replace(/\n/g, '<br>') + '</p>';
  });
  return parts.join('\n');
}
