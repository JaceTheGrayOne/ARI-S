// ARI.S Frontend JavaScript

// Import CSS for Vite HMR
import './style.css';

// Import Wails bindings
import { App, RetocService, UAssetService, RetocOperation, RetocResult, UAssetResult } from "../bindings/aris";
import { Events } from "@wailsio/runtime";

// Application state
let currentPane = 'home';
let currentOperationID = null;

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    
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
    initializeSettings();

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

    // Load saved settings (this will also apply the theme)
    loadSettings();
});

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
    
    // Removed open buttons as requested

    // TODO: Complete "Copy to /paks" feature - add game-paks-group div and choose-game-paks button to HTML
    // Game paks folder toggle
    // document.getElementById('copy-to-game-paks').addEventListener('change', (e) => {
    //     const group = document.getElementById('game-paks-group');
    //     group.style.display = e.target.checked ? 'flex' : 'none';
    // });

    // Choose game paks folder
    // document.getElementById('choose-game-paks').addEventListener('click', () => browseFolder('game-paks-folder'));

    // Action buttons
    document.getElementById('run-retoc').addEventListener('click', runRetocToZen);
    document.getElementById('run-unpak').addEventListener('click', runRetocUnpak);
    document.getElementById('cancel-unpak').addEventListener('click', cancelRetocOperation);
}

async function loadRetocSettings() {
    // Load saved paths and settings from App service
    const inputModFolder = await App.GetLastUsedPath('input_mod_folder') || '';
    const pakOutputDir = await App.GetLastUsedPath('pak_output_dir') || '';
    const gamePaksFolderUnpak = await App.GetLastUsedPath('game_paks_folder_unpak') || '';
    const extractOutputDir = await App.GetLastUsedPath('extract_output_dir') || '';
    const ueVersion = await App.GetPreference('ue_version') || 'UE5_4';

    document.getElementById('input-mod-folder').value = inputModFolder;
    document.getElementById('pak-output-dir').value = pakOutputDir;
    document.getElementById('game-paks-folder-unpak').value = gamePaksFolderUnpak;
    document.getElementById('extract-output-dir').value = extractOutputDir;
    document.getElementById('ue-version').value = ueVersion;
}

function runRetocToZen() {
    const inputPath = document.getElementById('input-mod-folder').value;
    const outputPath = document.getElementById('pak-output-dir').value;
    const ueVersion = document.getElementById('ue-version').value;
    const outputBaseName = document.getElementById('mod-name').value.trim();
    const orderSuffix = document.getElementById('load-order').value.trim();
    const usePriority = true; // Always use priority as requested
    // TODO: Complete "Copy to /paks" feature
    // const copyToGamePaks = document.getElementById('copy-to-game-paks').checked;
    // const gamePaksFolder = document.getElementById('game-paks-folder').value;
    const showConsole = true; // Always show console as requested

    if (!inputPath || !outputPath) {
        showOutput('retoc-output', 'Error: Please specify input and output paths.', 'error');
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

                // TODO: Complete "Copy to /paks" feature
                // if (copyToGamePaks && gamePaksFolder) {
                //     showOutput('retoc-output', `Copying files to: ${gamePaksFolder}`, 'info');
                //     // TODO: Implement file copying
                // }
            } else {
                showOutput('retoc-output', `Error: ${result.error}`, 'error');
                showOutput('retoc-output', result.output, 'error');
            }
        })
        .catch(error => {
            showOutput('retoc-output', `Error: ${error.message}`, 'error');
        });
}

function runRetocUnpak() {
    const gamePaksFolder = document.getElementById('game-paks-folder-unpak').value;
    const extractOutputDir = document.getElementById('extract-output-dir').value;
    const optionalSubfolder = ''; // Removed optional subfolder as requested
    const showConsole = true; // Always show console as requested

    if (!gamePaksFolder || !extractOutputDir) {
        showOutput('retoc-output', 'Error: Please specify game paks folder and extract output directory.', 'error');
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

    if (optionalSubfolder) {
        operation.options.push('--subfolder', optionalSubfolder);
    }

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
    
    // Action buttons
    document.getElementById('export-uassets').addEventListener('click', exportUAssets);
    document.getElementById('import-uassets').addEventListener('click', importUAssets);
    
    // File count updates
    document.getElementById('export-folder').addEventListener('input', updateExportFileCount);
    document.getElementById('import-folder').addEventListener('input', updateImportFileCount);
}

async function loadUAssetSettings() {
    const exportFolder = await App.GetLastUsedPath('export_folder') || '';
    const importFolder = await App.GetLastUsedPath('import_folder') || '';
    
    document.getElementById('export-folder').value = exportFolder;
    document.getElementById('import-folder').value = importFolder;
    
    updateExportFileCount();
    updateImportFileCount();
}

function exportUAssets() {
    const folderPath = document.getElementById('export-folder').value;
    
    if (!folderPath) {
        showOutput('uasset-output', 'Error: Please select a folder containing .uasset/.uexp files.', 'error');
        return;
    }
    
    // Save path to App service
    App.SetLastUsedPath('export_folder', folderPath);
    
    showOutput('uasset-output', 'Starting UAsset export to JSON...', 'info');
    
    UAssetService.ExportUAssets(folderPath)
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
    
    if (!folderPath) {
        showOutput('uasset-output', 'Error: Please select a folder containing .json files.', 'error');
        return;
    }
    
    // Save path to App service
    App.SetLastUsedPath('import_folder', folderPath);
    
    showOutput('uasset-output', 'Starting UAsset import from JSON...', 'info');
    
    UAssetService.ImportUAssets(folderPath)
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
        const [uassetCount, uexpCount, error] = await UAssetService.CountUAssetFiles(folderPath);

        if (error) {
            countElement.textContent = 'Error counting files';
            console.error('File counting error:', error);
        } else {
            countElement.textContent = `${uassetCount} .uasset, ${uexpCount} .uexp`;
        }
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
        const [jsonCount, error] = await UAssetService.CountJSONFiles(folderPath);

        if (error) {
            countElement.textContent = 'Error counting files';
            console.error('File counting error:', error);
        } else {
            countElement.textContent = `${jsonCount} .json`;
        }
    } catch (error) {
        countElement.textContent = 'Error counting files';
        console.error('File counting failed:', error);
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
    
    // Show notification
    const themeName = theme === 'light' ? 'Light' : 'Dark';
    showNotification(`Switched to ${themeName} theme`, 'success');
}

// Utility Functions
async function browseFolder(inputId) {
    try {
        const title = `Select folder for ${inputId.replace('-', ' ')}`;
        const result = await App.BrowseFolder(title);
        if (result) {
            document.getElementById(inputId).value = result;
        }
    } catch (error) {
        console.error('Error browsing folder:', error);
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

// Initialize drag and drop when DOM is ready
document.addEventListener('DOMContentLoaded', initializeDragAndDrop);