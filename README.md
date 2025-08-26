# 🎣 TwilightEvoFishing (Go + OpenGL)

A transparent, always-on-top overlay with interactive keyboard/mouse control that automates fishing mechanics in a game. Built in Go using OpenGL, GLFW, RobotGo, and screenshot detection. Supports multiple screen resolutions.

---

## ✨ Features

* ✅ Transparent overlay HUD using OpenGL
* 👡 Detects in-game visual cues (yellow, green, red, black pixels)
* 📊 Multi-resolution support (4K, 1440p, 1080p)
* 🧠 Automatically selects fishing spots based on Tier Selected
* 🐟 Detects hero death and item drops
* ⌨️ Hotkey Controls:

  * `Alt+F6` - Toggle ManaRefresh Off/12s/15s (Only works with slot 2 currently)
  * `Ctrl+F6` - Dodge Imp2 Arrows Toggle
  * `Ctrl+F8` - Toggle Fishing ON/OFF
  * `Ctrl+F7` - Cycle fishing Tiers

  * `Mouse Middle Click` - Can be configured to do two different things used for new resolution mapping
    * Print current mouse coordinates (for new resolution mapping)
    * Printscreen to identify scrolling combat(Like Dodge coordinate identification)

---

## 🛠️ Dependencies

This bot uses several native and Go-based libraries:

* [GoGL (gl & glfw)](https://github.com/go-gl)
* [gltext](https://github.com/go-gl/gltext)
* [robotgo](https://github.com/go-vgo/robotgo)
* [gohook](https://github.com/robotn/gohook)
* [screenshot](https://github.com/kbinani/screenshot)
* `golang.org/x/image/font/gofont/gobold`

---

## 🧰 Setup & Build

1. **Install Go 1.18+**

2. **Clone the repo:**

3. **Install dependencies:**

   ```bash
   go mod tidy
   ```

4. **Build and run:**

   Runs latest code with console (used for debugging or Resolution Mapping)
   ```bash
   go run main.go
   ```


   Builds executable file without console
   ```bash
   go build -ldflags="-H windowsgui" -o TwilightEveFishing.exe
   ```
   Run executable file that was created

---

## 🗄️ Supported Resolutions

Each resolution is mapped with specific coordinate sets:

* `3840x2160` (4K)
* `2560x1440` (1440p)
* `1920x1080` (1080p)

You can extend support for more resolutions by adding new entries to the `resolutions` map in `main.go`.

---

## 📋 Controls Summary

| Action                | Input                |
| --------------------- | -------------------- |
| Toggle Bot ON/OFF     | `Ctrl + F8`          |
| Change Fishing Tier   | `Ctrl + F7`          |
| Log Mouse Coordinates | `Middle Mouse Click` |

---

## 🧠 Tier System

Tiers correspond to different fishing locations and wait times.

```go
tiers := map[int]Tier{
  1: {WaitSec: 17},
  2: {WaitSec: 40},
  3: {WaitSec: 24},
  4: {WaitSec: 20},
  5: {WaitSec: 50},
}
```

---

## 🔧 Adding a New Resolution

To support a new screen resolution:

1. Add a new `Resolution` block in `resolutions`:

```go
"2560x1080": {
  Width: 2560, Height: 1080,
  MoveCheck: [2]int{...},
  HeroCheck: [2]int{...},
  YellowCheck: [2]int{...},
  GreenCheck: [2]int{...},
  InventorySlots: [][2]int{...},
  DropPoint: [2]int{...},
  FishingSpots: map[int][2]int{
    1: {...}, 2: {...}, ...
  },
},
```

2. Use `Middle Mouse Click` while in-game to log required coordinates.

---

## ⚠️ Safety Notes

This bot performs automated clicks and screen analysis. **Use responsibly** and only in accordance with the terms of service of the game you are interacting with. No guarantees of safety or compliance are provided.
