# Ebiten Radxa Dashboard Render Test

Standalone fake car dashboard / GPU stress test for Rock Pi 4B+ experiments. It does not use ALDL; it only tests whether Ebiten renders smoothly on your minimal display stack.

## Build On Radxa

Copy this folder to the Radxa, then run:

```sh
cd ebiten-radxa-dashboard-test
sh ./install-radxa-deps.sh
make build
./radxa-dashboard-test -fullscreen=true -stress=true
```

Or build and run in one command:

```sh
make run
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
cage ./radxa-dashboard-test -fullscreen=true -stress=true
```

Pure DRM/KMS without Wayland/X11 is not the normal Ebiten path, so assume you need X11 or Wayland.

## Controls

- `F`: fullscreen toggle
- `S`: stress toggle
- `H`: help toggle
- `Q` / `Esc`: quit

Watch `FPS`, `frame ms`, and the graph. Stable `60 FPS` and `~16.7ms` frames means good.
