import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0240 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-5 租户态平台治理动作隐藏', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-5a: tenant context cannot use platform menu or plugin governance actions', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0240();
  });
});
