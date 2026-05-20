import { test, expect } from '@host-tests/fixtures/auth';
import { smokeSourcePluginLifecycle } from '@host-tests/support/source-plugin-lifecycle';

test.describe('TC005 在线用户插件生命周期 smoke', () => {
  test('TC-1a: 在线用户插件可安装、启用、挂载菜单并访问页面', async ({
    adminPage,
  }) => {
    await smokeSourcePluginLifecycle(adminPage, {
      id: 'linapro-monitor-online',
      mountedTitles: ['在线用户'],
      route: '/monitor/online',
      assertAvailable: async (page) => {
        await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
        await expect(page.getByText(/在线用户列表/)).toBeVisible();
      },
    });
  });
});
