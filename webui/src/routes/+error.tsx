import { JSX } from 'solid-js'

interface ErrorProps {
  error: Error
  reset: () => void
}

export function Error(props: ErrorProps): JSX.Element {
  return (
    <div class="min-h-screen flex items-center justify-center bg-base-200 p-4">
      <div class="card bg-base-100 shadow-xl max-w-md w-full">
        <div class="card-body">
          <h2 class="card-title text-error">Error</h2>
          <p class="text-base-content/70">{props.error.message}</p>
          <div class="card-actions justify-end">
            <button 
              onClick={props.reset}
              class="btn btn-primary"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Error
