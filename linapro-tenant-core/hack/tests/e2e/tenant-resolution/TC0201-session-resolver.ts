import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0201 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-201 session 解析器', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-201a: switch flow persists tenant session and revokes the old one', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0201();
  });
});
