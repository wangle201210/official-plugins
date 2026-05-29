import type { APIRequestContext } from "@host-tests/support/playwright";

import { execFileSync } from "node:child_process";
import { rmSync } from "node:fs";
import path from "node:path";

import { test, expect } from "@host-tests/fixtures/auth";
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from "@host-tests/support/api/job";
import { waitForRouteReady } from "@host-tests/support/ui";
import { DemoDynamicPage } from "../../pages/DemoDynamicPage";

const pluginID = "linapro-demo-dynamic";
const sourcePluginID = "linapro-demo-source";
const pluginMenuNamePattern = /Dynamic Plugin Demo|动态插件示例/u;
const repoRoot = path.resolve(process.cwd(), "../..");
const legacyRuntimeArtifactPath = path.join(
  repoRoot,
  "apps",
  "lina-plugins",
  pluginID,
  "runtime",
  `${pluginID}.wasm`,
);

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;
let originalSourceInstalled = 0;
let originalSourceEnabled = 0;

function ensureRuntimePluginArtifact() {
  execFileSync("make", ["wasm", `p=${pluginID}`, "out=../../temp/output"], {
    cwd: repoRoot,
    stdio: "inherit",
  });
  rmSync(legacyRuntimeArtifactPath, { force: true });
}

async function ensurePluginInstalledAndEnabled() {
  await syncPlugins(adminApi);
  let sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  let plugin = await getPlugin(adminApi, pluginID);
  originalSourceInstalled = sourcePlugin.installed;
  originalSourceEnabled = sourcePlugin.enabled;
  originalInstalled = plugin.installed;
  originalEnabled = plugin.enabled;

  if (sourcePlugin.installed !== 1) {
    await installPlugin(adminApi, sourcePluginID, { installMode: "global" });
    sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  }
  if (sourcePlugin.enabled !== 1) {
    await enablePlugin(adminApi, sourcePluginID);
  }

  if (plugin.installed !== 1) {
    await installPlugin(adminApi, pluginID, { installMode: "global" });
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(adminApi, pluginID);
  }
}

async function restorePluginState() {
  let plugin = await getPlugin(adminApi, pluginID);

  if (originalInstalled !== 1) {
    if (plugin.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, pluginID);
    }
    await restoreSourcePluginState();
    return;
  }

  if (plugin.installed !== 1) {
    await installPlugin(adminApi, pluginID, { installMode: "global" });
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (originalEnabled === 1 && plugin.enabled !== 1) {
    await enablePlugin(adminApi, pluginID);
  } else if (originalEnabled !== 1 && plugin.enabled === 1) {
    await disablePlugin(adminApi, pluginID);
  }

  await restoreSourcePluginState();
}

async function restoreSourcePluginState() {
  let sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  if (originalSourceInstalled !== 1) {
    if (sourcePlugin.enabled === 1) {
      await disablePlugin(adminApi, sourcePluginID);
      sourcePlugin = await getPlugin(adminApi, sourcePluginID);
    }
    if (sourcePlugin.installed === 1) {
      await uninstallPlugin(adminApi, sourcePluginID);
    }
    return;
  }
  if (sourcePlugin.installed !== 1) {
    await installPlugin(adminApi, sourcePluginID, { installMode: "global" });
    sourcePlugin = await getPlugin(adminApi, sourcePluginID);
  }
  if (originalSourceEnabled === 1 && sourcePlugin.enabled !== 1) {
    await enablePlugin(adminApi, sourcePluginID);
  } else if (originalSourceEnabled !== 1 && sourcePlugin.enabled === 1) {
    await disablePlugin(adminApi, sourcePluginID);
  }
}

test.describe("TC-5 Manifest host service demo", () => {
  test.beforeAll(async () => {
    ensureRuntimePluginArtifact();
    adminApi = await createAdminApiContext();
    await ensurePluginInstalledAndEnabled();
  });

  test.afterAll(async () => {
    try {
      await restorePluginState();
    } finally {
      await adminApi.dispose();
    }
  });

  test("TC-5a: Manifest declaration is visible through the dynamic plugin page", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("English");
    await adminPage.reload({ waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage);

    const pluginPage = new DemoDynamicPage(adminPage);
    await pluginPage.clickSidebarMenuItem(pluginMenuNamePattern);
    await waitForRouteReady(adminPage);

    await expect(pluginPage.pluginDemoDynamicManifestDemo()).toBeVisible();
    await expect(pluginPage.pluginDemoDynamicManifestProfilePath()).toContainText(
      "config/profile.yaml",
    );
    await expect(pluginPage.pluginDemoDynamicManifestProfileName()).toContainText(
      "demo-dynamic-profile",
    );
    await expect(pluginPage.pluginDemoDynamicManifestConfigPath()).toContainText(
      "config/config.yaml",
    );
    await expect(pluginPage.pluginDemoDynamicManifestConfigPreview()).toContainText(
      "Hello from dynamic plugin",
    );
  });
});
