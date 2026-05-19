import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0178 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-178 多租户启用', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-178a: linapro-tenant-core plugin is installed and exposes real tenant APIs', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0178();
  });
});
