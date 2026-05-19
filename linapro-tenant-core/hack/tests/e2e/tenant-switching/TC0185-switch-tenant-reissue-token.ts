import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0185 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-185 切换租户重签 token', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-185a: switch-tenant reissues token and revokes the previous token', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0185();
  });
});
