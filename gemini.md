# Project Summary: VersaDumps Visualizer v2

This document summarizes the context and technical details of the `versaDumps` desktop application after its refactor to Vue.js and addition of new features.

## 1. Project Goal

The application is a desktop-based visualizer for debugging data, primarily from backend applications (e.g., PHP). It starts a local HTTP server that listens for incoming JSON payloads. These payloads, containing stack frame information (`file`, `line`, `function`) and a `context` object, are displayed in a real-time, interactive UI. The UI allows for easy exploration of nested data within the `context`.

Key UI features include:
- Light and Dark themes.
- Real-time log display with sorting controls (newest/oldest first).
- Ability to clear all logs or delete them individually.
- A dynamic window title that shows the number of received messages.

## 2. Technology Stack

*   **Backend Language**: Go
*   **Desktop App Framework**: Wails v2
*   **Frontend Framework**: **Vue.js 3**
*   **Frontend Build Tool**: Vite
*   **Configuration**: YAML (`config.yml`)
*   **Dependencies**:
    *   Go: `gopkg.in/yaml.v3`
    *   NPM: `vue`, `@vitejs/plugin-vue`

## 3. Project Structure

The project structure was refactored to support a Vue component-based architecture.

```
/versaDumps
├── app/
│   ├── build/bin/
│   │   ├── app.exe       # The final executable
│   │   └── config.yml  # Runtime configuration
│   ├── frontend/
│   │   ├── src/
│   │   │   ├── components/ # Reusable Vue components
│   │   │   │   ├── Icon.vue
│   │   │   │   ├── LogItem.vue
│   │   │   │   ├── JsonTreeView.vue
│   │   │   │   └── JsonTreeViewNode.vue
│   │   │   ├── App.vue       # Main Vue component
│   │   │   └── main.js     # Vue app initialization
│   │   ├── index.html
│   │   ├── package.json
│   │   └── vite.config.js
│   ├── app.go
│   ├── config.go
│   ├── go.mod
│   ├── main.go
│   └── server.go
└── gemini.md
```

## 4. Core Logic

### Data Structure

The application now expects a JSON payload with the following structure:

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
- `frame`: An object containing the source code location.
- `context`: A **string** containing a JSON object for detailed inspection.

### Backend (`.go` files)

- The core logic remains a bridge that passes data to the frontend.
- **New Feature**: It now maintains a counter for received messages and dynamically updates the application's window title via `runtime.WindowSetTitle`.

### Frontend (Vue.js)

- The frontend is a Single Page Application managed by Vue.
- **`App.vue`**: The main component holds the application state, including a reactive array of `logs`. It listens for the `newData` event from Wails and pushes new entries to the array.
- **State Management**: All display logic is reactive. For example, `sortedLogs` is a `computed` property that re-renders the list automatically when the sort order changes.
- **Component-Based UI**:
    - `LogItem.vue`: Displays a single log entry, including the new `frame` data, and handles the per-item delete logic by emitting an event to `App.vue`.
    - `JsonTreeView.vue` & `JsonTreeViewNode.vue`: Work together to recursively render the explorable JSON context, with local state (`isExpanded`) to manage expand/collapse.
- **Features Logic**:
    - **Theme:** A reactive `ref` in `App.vue` toggles a `data-theme` attribute on the main element. The choice is persisted to `localStorage`.
    - **Controls:** Buttons for sorting, clearing, and theme are icon-based and trigger methods in `App.vue` to change the state.

## 5. How to Build and Run

1.  **Development**: Navigate to the `app` directory and run `wails dev`. This provides hot-reloading and is best for making changes.
2.  **Production**: Navigate to the `app` directory and run `wails build`. This creates the final, self-contained `.exe` in the `app/build/bin` directory.
3.  **Execution**: Place `config.yml` in the same directory as `app.exe` and run the executable.