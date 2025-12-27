import { JSX } from 'solid-js'

export function Loading(): JSX.Element {
  return (
    <div class="min-h-screen flex items-center justify-center bg-base-200">
      <div class="flex flex-col items-center">
        <span class="loading loading-spinner loading-lg"></span>
        <p class="text-base-content/70 mt-4">Loading...</p>
      </div>
    </div>
  )
}

export default Loading
