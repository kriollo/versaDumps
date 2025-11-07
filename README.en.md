# VersaDumps Visualizer

**English** | [Español](README.md)

<p align="center">
  <img src="art/versaDumpsVisualizer.png" alt="VersaDumps Logo" />
</p>

## 📋 Description

VersaDumps Visualizer is a cross-platform desktop application for visualizing and debugging data in real-time, designed primarily for backend applications (such as PHP, Node.js, Python, etc.). The application starts a local HTTP server that listens for incoming JSON payloads and displays them in an interactive interface, in addition to monitoring system log files in real-time.

## ✨ Main Features

### 🎯 Core Features
- 🌓 **Light and dark themes** with persistent support
- 🌍 **Internationalization (i18n)**: Multi-language support (Spanish and English)
- 👤 **Multiple profiles**: Manage different configurations and switch between them easily
- 🔄 **Real-time visualization** of HTTP dumps with sorting controls (newest/oldest first)
- 📊 **Interactive exploration** of nested JSON data with expandable tree
- 📱 **Responsive design** with mobile-first approach

### 📂 Log File Monitoring
- 📁 **Real-time folder monitoring** with fsnotify
- 🔍 **Advanced filtering**:
  - By file extension (*.log, *.txt, etc.)
  - By log level (error, warning, info, debug, success)
  - By text content in real-time
- 📝 **Format support**: JSON and plain text
- 🔄 **Automatic log rotation detection**
- 🎨 **Syntax highlighting** for JSON files with differentiated colors
- 📊 **Line counter** for total and filtered lines
- 🗑️ **Complete management**: Clear logs, open folders, edit configuration

### 🖥️ User Interface
- 🔲 **Resizable split view**:
  - Top panel: HTTP Dumps (60% height by default)
  - Bottom panel: Log file viewer (40% height by default)
  - Adjustable separator with 30%-70% limits
- 🗑️ **Log management**: Clear all logs or delete them individually
- 📊 **Server status indicator** (online/offline/checking)
- 🔔 **Taskbar badge** (Windows) showing the number of messages received
- 🔍 **Dynamic window title** showing message counter
- 💾 **Saved window position** (restores size and position on restart)

### 🔄 Update System
- ✨ **Automatic updates** from GitHub Releases
- 📥 **Download and installation** of new versions
- 🔔 **Notifications** for available updates
- 📝 **Changelog** visible in the application

<p align="center">
  <img src="art/visualizerExample.png" alt="VersaDumps Example" />
</p>

### 📸 Screenshots

<div align="center">

**Profile Configuration**

<img src="art/visualizerExampleConfig1.png" alt="Configuration - Profiles" />

**Log Folder Management**

<img src="art/visualizerExampleConfig2.png" alt="Configuration - Log Folders" />

**Theme and Language Customization**

<img src="art/visualizerExampleConfig3.png" alt="Configuration - Theme and Language" />

</div>

## 🧰 Technology Stack

### Backend
- **Go** 1.21+
- **Wails v2** - Desktop application framework
- **fsnotify** - File system monitoring
- **yaml.v3** - Configuration handling

### Frontend
- **Vue.js 3** - Progressive JavaScript framework
- **Tailwind CSS** - Utility-first CSS framework
- **Vite** - Ultra-fast build tool

### Tools
- **PowerShell** - Automation and setup scripts
- **NSIS** - Windows installer
- **GitHub Actions** - Automated CI/CD

## 📁 Project Structure

```
/versaDumps
├── .github/
│   └── workflows/          # CI/CD pipelines
├── app/
│   ├── build/
│   │   ├── bin/
│   │   │   ├── VersaDumps.exe       # Final executable
│   │   │   └── config.yml           # Runtime configuration
│   │   └── appicon.png              # Application icon
│   ├── cmd/                         # CLI commands
│   ├── frontend/
│   │   ├── src/
│   │   │   ├── assets/              # Static resources
│   │   │   ├── components/          # Vue components
│   │   │   │   ├── ConfigModal.vue          # Configuration modal
│   │   │   │   ├── Icon.vue                 # Icon system
│   │   │   │   ├── JsonTreeView.vue         # JSON viewer
│   │   │   │   ├── JsonTreeViewNode.vue     # JSON tree nodes
│   │   │   │   ├── LineHr.vue               # Horizontal separator
│   │   │   │   ├── LogFileViewer.vue        # Log file viewer
│   │   │   │   ├── LogFoldersManager.vue    # Log folder manager
│   │   │   │   ├── LogItem.vue              # Individual log item
│   │   │   │   └── UpdateNotification.vue   # Update notification
│   │   │   ├── i18n/                # Internationalization
│   │   │   │   ├── en.js            # English translations
│   │   │   │   ├── es.js            # Spanish translations
│   │   │   │   └── index.js         # i18n configuration
│   │   │   ├── App.vue              # Main component
│   │   │   ├── index.css            # Global styles
│   │   │   └── main.js              # Entry point
│   │   ├── index.html
│   │   ├── package.json
│   │   ├── tailwind.config.js
│   │   └── vite.config.js
│   ├── tools/                       # Auxiliary tools
│   ├── app.go                       # Main app logic
│   ├── badge_windows.go             # Taskbar badge (Windows)
│   ├── badge_darwin.go              # Badge for macOS
│   ├── badge_unix.go                # Badge for Linux/Unix
│   ├── config.go                    # Configuration management
│   ├── config.yml                   # Configuration file
│   ├── logwatcher.go                # Log monitoring system
│   ├── main.go                      # Entry point
│   ├── server.go                    # HTTP server
│   ├── updater.go                   # Automatic update system
│   ├── updater_windows.go           # Update installer (Windows)
│   ├── updater_unix.go              # Update installer (Unix)
│   ├── go.mod
│   └── wails.json                   # Wails configuration
├── art/                             # Art resources
├── phpBack/                         # PHP integration example
│   └── composer.json                # versadumps-php package
├── test-logs/                       # Test logs
├── CHANGELOG.md                     # Change history
├── ICONS.md                         # Icon documentation
├── README.md                        # Spanish README
├── README.en.md                     # This file
├── RELEASE.md                       # Release notes
├── TODO.md                          # Task list
├── create-installer.ps1             # Installer creation script
├── release.ps1                      # Release script
├── setup-icons.ps1                  # Icon setup (Windows)
├── setup-icons.sh                   # Icon setup (Unix/macOS)
└── update-version.ps1               # Version update script
```

## 📊 Data Structure

### HTTP Payload

The application expects a JSON payload with the following structure:

```json
{
  "frame": {
    "file": "/path/to/file.php",
    "line": 123,
    "function": "myFunction"
  },
  "context": "{\"key\":\"value\"}"
}
```

- `frame`: Object containing the source code location
  - `file`: File path
  - `line`: Line number
  - `function`: Function name
- `context`: A **string** containing a JSON object for detailed inspection

### Configuration (config.yml)

```yaml
active_profile: Default
profiles:
  - name: Default
    server: localhost
    port: 9191
    theme: dark
    language: en
    show_types: true
    log_folders:
      - path: C:\logs\app
        extensions:
          - "*.log"
          - "*.txt"
        filters:
          - error
          - warning
        enabled: true
        format: text
      - path: C:\logs\json
        extensions:
          - "*.json"
        filters: []
        enabled: true
        format: json
  - name: Production
    server: 0.0.0.0
    port: 8080
    theme: light
    language: en
    show_types: false
    log_folders: []
window_position:
  x: 100
  y: 100
  width: 1200
  height: 800
```

#### Profile Configuration

- `active_profile`: Name of the active profile
- `profiles`: Array of configuration profiles
  - `name`: Profile name
  - `server`: HTTP server address (localhost, 0.0.0.0, etc.)
  - `port`: Port on which the server will listen
  - `theme`: Interface theme (`dark` or `light`)
  - `language`: Interface language (`es` or `en`)
  - `show_types`: Show data types in JSON viewer
  - `log_folders`: Log folders to monitor
    - `path`: Absolute folder path
    - `extensions`: File extensions to monitor (supports wildcards)
    - `filters`: Log level filters (empty = all)
    - `enabled`: Whether monitoring is active
    - `format`: Log format (`text` or `json`)

- `window_position`: Window position and size (optional)
  - `x`, `y`: Screen position
  - `width`, `height`: Window dimensions

## 🚀 How to Build and Run

### Prerequisites

- **Go** 1.21 or higher
- **Node.js** 16 or higher
- **Wails CLI** v2
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```

### Development

1. Navigate to the `app` directory
   ```bash
   cd app
   ```

2. Run Wails in development mode
   ```bash
   wails dev
   ```
   - Provides hot reload
   - Ideal for active development
   - Browser devtools available

### Production

1. Navigate to the `app` directory
   ```bash
   cd app
   ```

2. Build the application
   ```bash
   wails build
   ```
   - Creates the final standalone executable in `app/build/bin`

3. To create the Windows installer (from project root)
   ```powershell
   .\create-installer.ps1
   ```

### Running

1. Place `config.yml` in the same directory as the executable (automatically created with default values if it doesn't exist)
2. Run the executable file `VersaDumps.exe` (Windows) or `VersaDumps` (Linux/macOS)

## 🔧 Configuration

### Icon Setup

To customize the application icon:

1. Replace `app/build/appicon.png` with your icon (recommended: 256x256px PNG)
2. Run the setup script:
   ```powershell
   # On Windows
   .\setup-icons.ps1
   ```
   ```bash
   # On Unix/macOS/Linux
   ./setup-icons.sh
   ```
3. Rebuild the application with `wails build`

For more details, see [ICONS.md](ICONS.md).

### Profile Management

You can manage profiles directly from the interface:

1. Click on the settings icon (⚙️)
2. Select the profile you want to use or create a new one
3. Changes are automatically saved to `config.yml`

### Log Monitoring

To configure log folders:

1. Open the settings panel
2. Go to the "Log Folders" section
3. Add folders with their respective configurations
4. Logs will automatically appear in the bottom panel

## 🔌 PHP Integration

VersaDumps includes a PHP package to facilitate integration:

### Installation

```bash
composer require versadumps-php/versadumps-php
```

### Basic Usage

```php
<?php
require_once 'vendor/autoload.php';

use VersaDumps\VersaDumps;

// Configure the server (optional, default: localhost:9191)
VersaDumps::config([
    'host' => 'localhost',
    'port' => 9191
]);

// Dump data
$data = ['name' => 'John', 'age' => 30];
VersaDumps::dump($data);

// Dump with additional metadata
VersaDumps::dump($data, 'User processed');
```

For more information, see the [versadumps-php](https://github.com/kriollo/versadumps-php) repository.

## 💻 User Interface

### Main Components

- **App.vue**: Root component that manages global application state
- **LogItem.vue**: Displays an individual HTTP dump entry
- **JsonTreeView.vue** and **JsonTreeViewNode.vue**: Render explorable JSON context
- **LogFileViewer.vue**: Real-time log file viewer
- **LogFoldersManager.vue**: CRUD manager for log folders
- **ConfigModal.vue**: Configuration and profile management modal
- **UpdateNotification.vue**: Available update notification

### Icon System

The application includes a custom icon system:
- `gear` - Settings
- `trash` - Delete
- `sun` / `moon` - Theme toggle
- `sort` - Sort
- `file` - Files
- `edit` - Edit
- And more...

## 🔄 Update System

The application automatically checks for updates from GitHub:

- Checks on application startup
- Visual notification when an update is available
- One-click download and installation
- Changelog visible before updating
- Cross-platform support (Windows, macOS, Linux)

**Current version**: 3.0.1

## 🌍 Internationalization

Supported languages:
- 🇪🇸 Spanish (es)
- 🇬🇧 English (en)

You can change the language from the settings. Translations are dynamically loaded and saved in the active profile.

## 🎨 Themes

- **Dark Theme** (default): Ideal for development environments
- **Light Theme**: For different visual preferences

The theme is saved per profile and persists between sessions.

## 📦 Distribution

The application is distributed in two formats:

1. **NSIS Installer** (Windows): `versaDumps-installer-{version}.exe`
   - Guided installation
   - Start menu integration
   - Uninstaller included

2. **Portable executable**: `VersaDumps.exe` / `VersaDumps`
   - No installation required
   - Includes all dependencies
   - Cross-platform

## 🛠️ Development

### Go Code Structure

- `main.go`: Entry point, Wails initialization
- `app.go`: Application logic, state management
- `server.go`: HTTP server to receive dumps
- `logwatcher.go`: Log file monitoring system
- `config.go`: Configuration and profile management
- `updater.go`: Automatic update system
- `badge_*.go`: Platform-specific badge implementation

### Frontend Code Structure

- `App.vue`: Main application, global state management
- `components/`: Reusable Vue components
- `i18n/`: Translation system
- `assets/`: Static resources (images, fonts, etc.)

### Useful Scripts

```powershell
# Update version
.\update-version.ps1 -NewVersion "3.0.2"

# Create release
.\release.ps1

# Create installer
.\create-installer.ps1
```

## 🐛 Troubleshooting

### Server won't start

- Verify that the configured port is not in use
- Check the `config.yml` file
- Review application logs

### Logs are not updating

- Verify that the log folder exists
- Check read permissions
- Make sure extensions match your files
- Verify that monitoring is enabled in settings

### Badge doesn't appear on Windows

- Requires Windows 7 or higher
- Verify that the application has appropriate permissions

## 🤝 Contributing

Contributions are welcome. Please:

1. Fork the project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

Please make sure to:
- Follow code best practices
- Add tests for new features
- Update corresponding documentation
- Maintain backward compatibility when possible

## 📄 License

This project is licensed under the [MIT License](LICENSE).

## 🔗 Useful Links

- [Main repository](https://github.com/kriollo/versaDumps)
- [PHP package](https://github.com/kriollo/versadumps-php)
- [Releases](https://github.com/kriollo/versaDumps/releases)
- [Issues](https://github.com/kriollo/versaDumps/issues)
- [Changelog](CHANGELOG.md)

## 👨‍💻 Author

**kriollo**
- Email: kriollone@gmail.com
- GitHub: [@kriollo](https://github.com/kriollo)

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/kriollo">kriollo</a>
</p>
