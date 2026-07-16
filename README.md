# Ebiten Radxa Dashboard Test

Standalone fake car dashboard for Rock Pi 4B+ experiments. It does not use ALDL/CAN yet; keyboard input feeds a clean `VehicleState` that the renderer draws.

## Build On Radxa

Install the native packages Ebiten needs, then build on the Radxa:

```sh
cd ebiten-radxa-dashboard-test
sudo apt update
sudo apt install -y build-essential libgl1-mesa-dev libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev libasound2-dev
make build
./radxa-dashboard-test -fullscreen=true
```

For a quick windowed local run:

```sh
make run
```

If cross-compiling from another Linux ARM64-capable setup with the right C toolchain available:

```sh
make build-radxa
```

## Minimal X11 Kiosk Test

Ebiten on Linux is safest with X11. For a minimal boot test, use `startx`/Openbox first:

```sh
startx ./radxa-dashboard-test -- -nocursor
```

Wayland + Cage might work only if this Ebiten/GLFW build runs as a native Wayland client on your image. If not, use X11 for now. Do not make Cage the first test.

## Display Stack

Ebiten needs a display backend. For lowest-pain testing use Wayland + Cage or X11. Cage is a good future path for kiosk boot:

```sh
cage ./radxa-dashboard-test -fullscreen=true
```

Pure DRM/KMS without Wayland/X11 is not the normal Ebiten path, so assume you need X11 or Wayland.

## Controls

- `F`: fullscreen toggle
- `Up` / `Down`: change fake speed and RPM
- `Left` / `Right`: change fake gear
- `Space`: warning on
- `C`: warning off
- `Q` / `Esc`: quit

The important architecture is `input/source -> VehicleState -> Draw`. Later, replace `KeyboardVehicleSource` with a CAN/OBD/replay source without making the drawing code parse raw vehicle messages.
