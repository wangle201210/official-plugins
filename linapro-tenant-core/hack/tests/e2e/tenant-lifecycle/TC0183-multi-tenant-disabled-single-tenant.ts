import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0183 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-183 多租户禁用退化', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-183a: linapro-tenant-core lifecycle precondition keeps the platform-only plugin installed', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0183();
  });
});
