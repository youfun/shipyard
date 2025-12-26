defmodule DeployExamplePhoenixWeb.TodoLive.Index do
  use DeployExamplePhoenixWeb, :live_view

  alias DeployExamplePhoenix.Todos

  @impl true
  def render(assigns) do
    ~H"""
    <Layouts.app flash={@flash}>
      <.header>
        Listing Todos
        <:actions>
          <.button variant="primary" navigate={~p"/todos/new"}>
            <.icon name="hero-plus" /> New Todo
          </.button>
        </:actions>
      </.header>

      <.table
        id="todos"
        rows={@streams.todos}
        row_click={fn {_id, todo} -> JS.navigate(~p"/todos/#{todo}") end}
      >
        <:col :let={{_id, todo}} label="Title">{todo.title}</:col>
        <:col :let={{_id, todo}} label="Description">{todo.description}</:col>
        <:col :let={{_id, todo}} label="Completed">{todo.completed}</:col>
        <:action :let={{_id, todo}}>
          <div class="sr-only">
            <.link navigate={~p"/todos/#{todo}"}>Show</.link>
          </div>
          <.link navigate={~p"/todos/#{todo}/edit"}>Edit</.link>
        </:action>
        <:action :let={{id, todo}}>
          <.link
            phx-click={JS.push("delete", value: %{id: todo.id}) |> hide("##{id}")}
            data-confirm="Are you sure?"
          >
            Delete
          </.link>
        </:action>
      </.table>
    </Layouts.app>
    """
  end

  @impl true
  def mount(_params, _session, socket) do
    {:ok,
     socket
     |> assign(:page_title, "Listing Todos")
     |> stream(:todos, list_todos())}
  end

  @impl true
  def handle_event("delete", %{"id" => id}, socket) do
    todo = Todos.get_todo!(id)
    {:ok, _} = Todos.delete_todo(todo)

    {:noreply, stream_delete(socket, :todos, todo)}
  end

  defp list_todos() do
    Todos.list_todos()
  end
end
