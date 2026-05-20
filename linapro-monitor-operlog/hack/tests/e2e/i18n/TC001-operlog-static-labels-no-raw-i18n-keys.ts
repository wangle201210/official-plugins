import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { waitForTableReady } from '@host-tests/support/ui';

const untranslatedKeyPattern = /\b(?:plugin|pages)\.[A-Za-z0-9_.:-]+\b/g;

test.describe('TC001 操作日志静态文案不再泄漏原始 i18n key', () => {
  test('TC-1a: 英文环境下操作日志页展示翻译后的静态文案', async ({
    adminPage,
    mainLayout,
  }) => {
    await ensureSourcePluginEnabled(adminPage, 'linapro-monitor-operlog');
    await mainLayout.switchLanguage('English');

    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    const bodyText = await adminPage.locator('body').innerText();
    expect(bodyText).toContain('Module Name');
    expect(bodyText).toContain('Actions');
    expect([...new Set(bodyText.match(untranslatedKeyPattern) || [])]).toEqual([]);
  });

  test('TC-1b: 运行时语言包接口返回操作日志业务 i18n 资源', async ({
    adminPage,
  }) => {
    const response = await adminPage.request.get(
      '/api/v1/i18n/runtime/messages?lang=en-US',
      {
        headers: {
          'Accept-Language': 'en-US',
        },
      },
    );
    expect(response.ok()).toBeTruthy();
    const payload = await response.json();
    const messages = payload?.data?.messages ?? payload?.messages;

    expect(messages?.plugin?.['linapro-monitor-operlog']?.fields?.moduleName).toBe(
      'Module Name',
    );
  });
});
