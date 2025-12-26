import { For, Show, JSX } from 'solid-js'
import { useI18n } from '@i18n'
import type { Domain } from '@types'

export function DomainTab(props: { domains: Domain[]; isLoading: boolean }): JSX.Element {
  const { t } = useI18n()

  return (
    <div>
      <Show when={!props.isLoading} fallback={
        <div class="flex justify-center py-8">
          <span class="loading loading-spinner loading-md"></span>
        </div>
      }>
        <Show when={props.domains.length > 0} fallback={
          <div class="text-center py-8 text-base-content/50">
            No domains configured
          </div>
        }>
          <table class="table">
            <thead>
              <tr>
                <th>Domain</th>
                <th>Primary</th>
                <th>Created</th>
              </tr>
            </thead>
            <tbody>
              <For each={props.domains}>
                {(domain) => (
                  <tr class="hover">
                    <td>{domain.domainName}</td>
                    <td>
                      <span classList={{
                        'badge': true,
                        'badge-primary': domain.isActive,
                        'badge-ghost': !domain.isActive,
                      }}>
                        {domain.isActive ? 'Yes' : 'No'}
                      </span>
                    </td>
                    <td>{domain.createdAt}</td>
                  </tr>
                )}
              </For>
            </tbody>
          </table>
        </Show>
      </Show>
    </div>
  )
}
