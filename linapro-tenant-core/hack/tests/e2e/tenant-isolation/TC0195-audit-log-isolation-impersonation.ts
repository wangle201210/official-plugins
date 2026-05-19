import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0195 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-195 审计日志隔离与 impersonation 标记', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-195a: login and operation logs keep tenant and impersonation fields', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0195();
  });
});
