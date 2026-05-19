import { test, expect } from '../../support/linapro-tenant-core';
import { scenarioTC0211 } from '../../support/linapro-tenant-core-scenarios';

test.describe('TC-211 force 通道审计', () => {
  test.use({ multiTenantMode: 'linapro-tenant-core-enabled' });

  test('TC-211a: force uninstall request stays governed and leaves plugin installed', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('linapro-tenant-core-enabled');
    await scenarioTC0211();
  });
});
