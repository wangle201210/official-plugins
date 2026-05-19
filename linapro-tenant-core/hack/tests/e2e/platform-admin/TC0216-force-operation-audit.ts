import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0216 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-216 强制操作审计', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-216a: force lifecycle request is protected and observable through stable state', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0216();
  });
});
