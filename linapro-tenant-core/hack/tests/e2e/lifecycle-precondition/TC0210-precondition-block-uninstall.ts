import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0210 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-210 Lifecycle Precondition 否决卸载', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-210a: lifecycle precondition blocks linapro-tenant-core uninstall', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0210();
  });
});
