import { test, expect } from '@host-tests/fixtures/auth';
import { ensureSourcePluginEnabled } from '@host-tests/fixtures/plugin';
import { waitForTableReady } from '@host-tests/support/ui';

const untranslatedKeyPattern = /\b(?:plugin|pages)\.[A-Za-z0-9_.:-]+\b/g;

test.describe('TC002 登录日志静态文案不再泄漏原始 i18n key', () => {
  test('TC-2a: 英文环境下登录日志页展示翻译后的静态文案', async ({
    adminPage,
    mainLayout,
  }) => {
    await ensureSourcePluginEnabled(adminPage, 'linapro-monitor-loginlog');
    await mainLayout.switchLanguage('English');

    await adminPage.goto('/monitor/loginlog');
    await waitForTableReady(adminPage);

    const bodyText = await adminPage.locator('body').innerText();
    expect(bodyText).toContain('User Account');
    expect(bodyText).toContain('Actions');
    expect([...new Set(bodyText.match(untranslatedKeyPattern) || [])]).toEqual([]);
  });

  test('TC-2b: 运行时语言包接口返回登录日志业务 i18n 资源', async ({
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

    expect(messages?.plugin?.['linapro-monitor-loginlog']?.fields?.userName).toBe(
      'User Account',
    );
  });
});
