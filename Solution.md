# Solutions for Flutter Web WASM Jitter on RHEL (Chrome)

The "jitter" or stuttering performance experienced specifically on RHEL when running the Flutter Web WASM application is almost certainly caused by **Hardware Acceleration (WebGL) being blocked or disabled by default** in Chrome.

RHEL prioritizes extreme stability and uses conservative graphics driver policies, often placing GPU drivers on a "blocklist." This forces the browser to use software rendering (calculating the entire UI directly on the CPU via SwiftShader), which results in poor performance and jitter.

Here are the steps to diagnose and fix this issue:

### 1. Verify Software Extrapolation (The "Jitter" Engine)

First, confirm if Chrome is falling back to software rendering:

1. Open a new tab in Chrome on the RHEL system.
2. Navigate to `chrome://gpu`.
3. Look under the **"Graphics Feature Status"** section.
4. If you see items like `WebGL: Software only, hardware acceleration unavailable`, then Chrome is rendering the Flutter app using only the CPU.

### 2. Force Enable Hardware Acceleration

To bypass RHEL's conservative driver blocklist and force Chrome to use your graphics hardware directly:

1. Open Chrome and navigate to `chrome://flags`.
2. Search for **"Override software rendering list"** (or `#ignore-gpu-blocklist`).
3. Set the flag to **Enabled**.
4. Relaunch Chrome and test the application.

### 3. Wayland vs X11 Check (Common on RHEL 8/9)

RHEL defaults to the Wayland display server protocol. By default, Chrome might run through an emulation layer (Xwayland) without direct hardware acceleration, causing massive stutter in heavy WASM WebGL apps.

To force Chrome to run natively on Wayland with full GPU support:

1. Open Chrome and navigate to `chrome://flags`.
2. Search for **"Preferred Ozone platform"** (or `#ozone-platform-hint`).
3. Set the flag from `Default` to **Wayland**.
4. Relaunch Chrome.

### 4. Quick Terminal Test

Alternatively, you can quickly test all these flags at once by launching Chrome directly from the terminal with the required arguments. This will instantly show if the jitter disappears without changing permanent Chrome settings:

```bash
google-chrome --ignore-gpu-blocklist --enable-features=Vulkan --ozone-platform-hint=wayland
```

Once hardware acceleration is successfully enabled and running natively on the RHEL system, the Flutter WASM engine will smoothly utilize WebGL, reaching the 60fps performance seen on Ubuntu and Fedora.
