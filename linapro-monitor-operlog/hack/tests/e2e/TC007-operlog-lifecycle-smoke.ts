import { test, expect } from '@host-tests/fixtures/auth';
import { smokeSourcePluginLifecycle } from '@host-tests/support/source-plugin-lifecycle';

test.describe('TC007 操作日志插件生命周期 smoke', () => {
  test('TC-3a: 操作日志插件可安装、启用、挂载菜单并访问页面', async ({
    adminPage,
  }) => {
    await smokeSourcePluginLifecycle(adminPage, {
      id: 'linapro-monitor-operlog',
      mountedTitles: ['操作日志'],
      route: '/monitor/operlog',
      assertAvailable: async (page) => {
        await expect(page.locator('.vxe-table')).toBeVisible({ timeout: 10000 });
        await expect(page.getByRole('button', { name: /清\s*空/ })).toBeVisible();
      },
    });
  });
});
