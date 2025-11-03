// ARI.S Frontend JavaScript

// Import CSS for Vite HMR
import './style.css';

// Import Wails bindings
import { App } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/app";
import { RetocService, RetocOperation, RetocResult } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/retoc";
import { UAssetService, UAssetResult } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/uasset";
import { InjectorService } from "../bindings/github.com/JaceTheGrayOne/ARI-S/internal/injector";
import { Events } from "@wailsio/runtime";

// Application state
let currentPane = 'home';

// Track which output paths were explicitly set by user (not just loaded from config)
let userSetOutputPaths = {
    pak_output_dir: false,
    extract_output_dir: false
};

// Track if settings have been loaded for the first time (to show console logs only on app launch)
let firstLoad = {
    retoc: true,
    uasset: true
};

// Initialize the application
document.addEventListener('DOMContentLoaded', async function() {

    // Test backend connection
    console.log('Testing backend connection...');
    try {
        App.GetPreference('test').then(result => {
            console.log('Backend connection successful:', result);
        }).catch(error => {
            console.error('Backend connection failed:', error);
        });
    } catch (error) {
        console.error('Backend service not available:', error);
    }

    // Initialize UI
    initializeNavigation();
    initializeRetocManager();
    initializeUAssetManager();
    initializeInjectorManager();
    initializeSettings();
    initializeDragAndDrop();

    // Set up event listeners for retoc events
    Events.On("retoc:started", (data) => {
        currentOperationID = data.operation_id;
        console.log(`Operation started with ID: ${currentOperationID}`);
    });

    Events.On("retoc:output", (data) => {
        if (data.operation_id === currentOperationID) {
            showOutput('retoc-output', data.line, 'info');
        }
    });

    // Load initial configuration from backend (loaded at app startup)
    // This populates all fields at once before user interacts with UI
    try {
        const initialConfig = await App.GetInitialConfig();
        await populateFieldsFromConfig(initialConfig);
        console.log('Initial configuration loaded and applied');
    } catch (error) {
        console.error('Failed to load initial configuration:', error);
    }

    // Load and apply theme settings
    loadSettings();
});

// Populate all form fields from the initial configuration
async function populateFieldsFromConfig(config) {
    // Retoc pane fields
    if (config.input_mod_folder) {
        document.getElementById('input-mod-folder').value = config.input_mod_folder;
    }
    if (config.pak_output_dir) {
        document.getElementById('pak-output-dir').value = config.pak_output_dir;
        // Validate and mark as user-set if valid
        if (await App.ValidateDirectory(config.pak_output_dir)) {
            userSetOutputPaths.pak_output_dir = true;
        }
    }
    if (config.game_paks_folder_unpak) {
        document.getElementById('game-paks-folder-unpak').value = config.game_paks_folder_unpak;
    }
    if (config.extract_output_dir) {
        document.getElementById('extract-output-dir').value = config.extract_output_dir;
        // Validate and mark as user-set if valid
        if (await App.ValidateDirectory(config.extract_output_dir)) {
            userSetOutputPaths.extract_output_dir = true;
        }
    }

    // UAsset pane fields
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

    // Injector pane fields
    if (config.dll_path) {
        document.getElementById('dll-path').value = config.dll_path;
    }

    // Settings
    if (config.ue_version) {
        document.getElementById('ue-version').value = config.ue_version;
    }
    if (config.theme) {
        document.getElementById('theme-select').value = config.theme;
    }
    if (config.auto_save) {
        document.getElementById('auto-save-settings').checked = config.auto_save === 'true';
    }
}

// Navigation Management
function initializeNavigation() {
    const navItems = document.querySelectorAll('.nav-item');
    const tiles = document.querySelectorAll('.tile');
    
    // Handle sidebar navigation
    navItems.forEach(item => {
        item.addEventListener('click', () => {
            const pane = item.dataset.pane;
            switchPane(pane);
        });
    });
    
    // Handle tile navigation
    tiles.forEach(tile => {
        tile.addEventListener('click', () => {
            const pane = tile.dataset.pane;
            switchPane(pane);
        });
    });
}

// New helper for the staggered fade/slide reveals
function animatePaneElements(pane) {
  const fadeEls = pane.querySelectorAll('.reveal-fade, .reveal-scale');
  fadeEls.forEach((el, i) => {
    el.style.animationDelay = `${i * 0.10}s`; // slightly slower reveal for deeper elements
    // Restart the animation each time we enter a pane
    el.style.animation = 'none';
    void el.offsetWidth; // force reflow to reset animation
    el.style.animation = '';
  });
}

// Replace your old switchPane() with this new version
function switchPane(paneName) {
  // Update navigation
  document.querySelectorAll('.nav-item').forEach(item => {
    item.classList.remove('active');
  });
  document.querySelector(`[data-pane="${paneName}"]`)?.classList.add('active');

  // Update panes
  document.querySelectorAll('.pane').forEach(pane => {
    pane.classList.remove('active');
  });
  const next = document.getElementById(`${paneName}-pane`);
  if (!next) return;
  next.classList.add('active');

  currentPane = paneName;

  // Load pane-specific data (keep your existing logic)
  if (paneName === 'retoc') {
    loadRetocSettings();
  } else if (paneName === 'uasset') {
    loadUAssetSettings();
  } else if (paneName === 'injector') {
    loadProcessList();
  }

  // Trigger staggered element reveals after the pane becomes active
  animatePaneElements(next);
}


// Retoc Manager
function initializeRetocManager() {
    // Browse buttons
    document.getElementById('browse-input-mod').addEventListener('click', () => browseFolder('input-mod-folder'));
    document.getElementById('browse-pak-output').addEventListener('click', () => browseFolder('pak-output-dir'));
    document.getElementById('browse-game-paks-unpak').addEventListener('click', () => browseFolder('game-paks-folder-unpak'));
    document.getElementById('browse-extract-output').addEventListener('click', () => browseFolder('extract-output-dir'));

    // Mark output paths as user-set when manually typed/changed
    document.getElementById('pak-output-dir').addEventListener('input', () => {
        userSetOutputPaths.pak_output_dir = true;
    });
    document.getElementById('extract-output-dir').addEventListener('input', () => {
        userSetOutputPaths.extract_output_dir = true;
    });

    // Action buttons
    document.getElementById('run-retoc').addEventListener('click', runRetocToZen);
    document.getElementById('run-unpak').addEventListener('click', runRetocUnpak);
    document.getElementById('cancel-unpak').addEventListener('click', cancelRetocOperation);
}

async function loadRetocSettings() {
    // Fields are already populated by populateFieldsFromConfig on app startup
    // This function now only handles console logging on first pane visit

    // Log loaded paths to console only on first visit to this pane
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

    // Validate that output path was explicitly set by user
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

    // Validate serialization number (only digits 0-9)
    if (orderSuffix && !/^\d+$/.test(orderSuffix)) {
        showOutput('retoc-output', 'Error: Serialization number must contain only digits 0-9.', 'error');
        return;
    }

    // Validate mod name (no Windows invalid filename characters)
    // Invalid characters: < > : " / \ | ? *
    if (outputBaseName && /[<>:"/\\|?*]/.test(outputBaseName)) {
        showOutput('retoc-output', 'Error: Mod name contains invalid characters. Cannot use: < > : " / \\ | ? *', 'error');
        return;
    }

    // Save paths to App service
    App.SetLastUsedPath('input_mod_folder', inputPath);
    App.SetLastUsedPath('pak_output_dir', outputPath);
    App.SetPreference('ue_version', ueVersion);

    showOutput('retoc-output', 'Starting Retoc to-zen operation...', 'info');

    // Build operation - output_path should be directory only for to-zen
    const operation = new RetocOperation({
        command: 'to-zen',
        input_path: inputPath,
        output_path: outputPath,  // Directory only, not filename
        ue_version: ueVersion,
        options: []
    });

    // Add mod name and serialization for file renaming (backend will use defaults if not provided)
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

    // Validate that output path was explicitly set by user
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

    // Save paths to App service
    App.SetLastUsedPath('game_paks_folder_unpak', gamePaksFolder);
    App.SetLastUsedPath('extract_output_dir', extractOutputDir);

    // Show cancel button, hide execute button
    document.getElementById('run-unpak').style.display = 'none';
    document.getElementById('cancel-unpak').style.display = 'inline-block';

    showOutput('retoc-output', 'Starting Retoc extraction operation (to-legacy)...', 'info');

    // Build operation - use 'to-legacy' instead of 'unpack'
    const operation = new RetocOperation({
        command: 'to-legacy',
        input_path: gamePaksFolder,
        output_path: extractOutputDir,
        options: []
    });

    // Execute operation - since this is a blocking call, we can't get the operation ID
    // until it completes. We'll use CancelCurrentOperation instead.
    RetocService.RunRetoc(operation)
        .then(result => {
            // Hide cancel button, show execute button
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
            // Hide cancel button, show execute button
            document.getElementById('run-unpak').style.display = 'inline-block';
            document.getElementById('cancel-unpak').style.display = 'none';
            currentOperationID = null;

            showOutput('retoc-output', `Error: ${error.message}`, 'error');
        });
}

function cancelRetocOperation() {
    showOutput('retoc-output', 'Cancelling operation...', 'warning');

    // Cancel the to-legacy extraction operation specifically
    RetocService.CancelOperationByCommand('to-legacy')
        .then(() => {
            showOutput('retoc-output', 'Cancellation requested', 'info');
        })
        .catch(error => {
            showOutput('retoc-output', `Failed to cancel: ${error.message}`, 'error');
        });
}

// Preview functions removed as requested

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
    // Fields are already populated by populateFieldsFromConfig on app startup
    // This function now only handles console logging on first pane visit

    // Log loaded paths to console only on first visit to this pane
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
}

function exportUAssets() {
    const folderPath = document.getElementById('export-folder').value;
    const mappingsPath = document.getElementById('uasset-mappings-path').value;

    if (!folderPath) {
        showOutput('uasset-output', 'Error: Please select a folder containing .uasset/.uexp files.', 'error');
        return;
    }

    // Save paths to App service
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

    // Save paths to App service
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
        // CountUAssetFiles returns [uassetCount, uexpCount] tuple
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
        // CountJSONFiles returns a number, not an array
        const jsonCount = await UAssetService.CountJSONFiles(folderPath);
        countElement.textContent = `${jsonCount} .json`;
    } catch (error) {
        countElement.textContent = 'Error counting files';
        console.error('File counting failed:', error);
    }
}

// DLL Injector Manager
function initializeInjectorManager() {
    // Browse button for DLL file
    document.getElementById('browse-dll').addEventListener('click', () => browseDLL('dll-path'));

    // Refresh processes button
    document.getElementById('refresh-processes').addEventListener('click', loadProcessList);

    // Process selector change handler
    document.getElementById('process-selector').addEventListener('change', () => {
        const selector = document.getElementById('process-selector');
        const processInfo = document.getElementById('process-info');
        const selectedPidElement = document.getElementById('selected-pid');

        if (selector.value) {
            const pid = selector.value;
            selectedPidElement.textContent = pid;
            processInfo.style.display = 'block';
        } else {
            processInfo.style.display = 'none';
        }
    });

    // Inject button
    document.getElementById('inject-dll').addEventListener('click', injectDLL);

    // Load process list on pane activation
    // This will be triggered when switching to injector pane
}

async function browseDLL(inputId) {
    try {
        const title = 'Select DLL file to inject';
        const filter = "DLL Files\x00*.dll\x00All Files\x00*.*\x00\x00";
        const key = inputId.replace(/-/g, '_');

        const result = await App.BrowseFile(title, filter, key);
        if (result) {
            document.getElementById(inputId).value = result;
            await App.SetLastUsedPath(key, result);
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
        const processes = await InjectorService.GetRunningProcesses();
        const selector = document.getElementById('process-selector');

        // Clear existing options except the first one
        selector.innerHTML = '<option value="">-- Select a process --</option>';

        // Sort processes by name
        processes.sort((a, b) => a.name.localeCompare(b.name));

        // Add process options
        processes.forEach(proc => {
            const option = document.createElement('option');
            option.value = proc.pid;
            option.textContent = `${proc.name} (PID: ${proc.pid})`;
            selector.appendChild(option);
        });

        showOutput('injector-output', `Loaded ${processes.length} running processes`, 'success');
    } catch (error) {
        console.error('Error loading processes:', error);
        showOutput('injector-output', `Error loading processes: ${error.message}`, 'error');
    }
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
            // Check if this is an elevation request
            if (result.error === 'NEEDS_ELEVATION') {
                showOutput('injector-output', '⚠️ ADMINISTRATOR PRIVILEGES REQUIRED', 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', result.output, 'info');
                showOutput('injector-output', '---', 'info');
                showOutput('injector-output', 'Your DLL path has been saved.', 'success');
                showOutput('injector-output', 'After restart, simply click "Inject DLL" again.', 'info');
            } else {
                // Regular error
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
        // Re-enable inject button
        injectButton.disabled = false;
        injectButton.textContent = 'Inject DLL';
    }
}

// Settings
function initializeSettings() {
    document.getElementById('save-settings').addEventListener('click', saveSettings);
    document.getElementById('theme-select').addEventListener('change', (e) => {
        applyTheme(e.target.value);
    });
}

async function loadSettings() {
    const theme = await App.GetPreference('theme') || 'dark';
    const autoSave = await App.GetPreference('auto_save') === 'true';
    
    document.getElementById('theme-select').value = theme;
    document.getElementById('auto-save-settings').checked = autoSave;
    
    // Apply the loaded theme
    applyTheme(theme);
}

function saveSettings() {
    const theme = document.getElementById('theme-select').value;
    const autoSave = document.getElementById('auto-save-settings').checked;
    
    App.SetPreference('theme', theme);
    App.SetPreference('auto_save', autoSave.toString());
    
    showNotification('Settings saved successfully!', 'success');
}

function applyTheme(theme) {
    const body = document.body;
    
    // Remove existing theme classes
    body.classList.remove('light-theme', 'dark-theme');
    
    // Apply the new theme
    if (theme === 'light') {
        body.classList.add('light-theme');
    } else {
        body.classList.add('dark-theme');
    }
    
    // Save the theme preference
    App.SetPreference('theme', theme);
}

// Utility Functions
async function browseFolder(inputId) {
    try {
        const title = `Select folder for ${inputId.replace('-', ' ')}`;
        const key = inputId.replace(/-/g, '_'); // Convert input-mod-folder to input_mod_folder

        // Pass the key so BrowseFolder can remember the last location for this specific field
        const result = await App.BrowseFolder(title, key);
        if (result) {
            document.getElementById(inputId).value = result;

            // Mark output paths as user-set when browsed
            if (key === 'pak_output_dir' || key === 'extract_output_dir') {
                userSetOutputPaths[key] = true;
            }

            // Trigger file count updates for UAsset folders
            if (inputId === 'export-folder') {
                updateExportFileCount();
            } else if (inputId === 'import-folder') {
                updateImportFileCount();
            }

            // Save the path for this specific field to remember location for next time
            await App.SetLastUsedPath(key, result);
        }
    } catch (error) {
        console.error('Error browsing folder:', error);
    }
}

async function browseFile(inputId) {
    try {
        const title = `Select file for ${inputId.replace('-', ' ')}`;
        const filter = "USMAP Files\x00*.usmap\x00All Files\x00*.*\x00\x00"; // Default filter

        // Convert input ID to config key (e.g., "uasset-mappings-path" -> "uasset_mappings_path")
        const key = inputId.replace(/-/g, '_');

        const result = await App.BrowseFile(title, filter, key);
        if (result) {
            document.getElementById(inputId).value = result;

            // Save the selected path for next time
            await App.SetLastUsedPath(key, result);
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
}

function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    
    // Style the notification
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
    
    // Set background based on type
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
    
    // Add to page
    document.body.appendChild(notification);
    
    // Animate in
    setTimeout(() => {
        notification.style.opacity = '1';
        notification.style.transform = 'translateX(0)';
    }, 10);
    
    // Remove after 3 seconds
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

// Drag and Drop Support
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