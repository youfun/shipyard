defmodule DeployExamplePhoenixWeb.TodoLive.Form do
  use DeployExamplePhoenixWeb, :live_view

  alias DeployExamplePhoenix.Todos
  alias DeployExamplePhoenix.Todos.Todo

  @impl true
  def render(assigns) do
    ~H"""
    <Layouts.app flash={@flash}>
      <.header>
        {@page_title}
        <:subtitle>Use this form to manage todo records in your database.</:subtitle>
      </.header>

      <.form for={@form} id="todo-form" phx-change="validate" phx-submit="save">
        <.input field={@form[:title]} type="text" label="Title" />
        <.input field={@form[:description]} type="textarea" label="Description" />
        <.input field={@form[:completed]} type="checkbox" label="Completed" />
        <footer>
          <.button phx-disable-with="Saving..." variant="primary">Save Todo</.button>
          <.button navigate={return_path(@return_to, @todo)}>Cancel</.button>
        </footer>
      </.form>
    </Layouts.app>
    """
  end

  @impl true
  def mount(params, _session, socket) do
    {:ok,
     socket
     |> assign(:return_to, return_to(params["return_to"]))
     |> apply_action(socket.assigns.live_action, params)}
  end

  defp return_to("show"), do: "show"
  defp return_to(_), do: "index"

  defp apply_action(socket, :edit, %{"id" => id}) do
    todo = Todos.get_todo!(id)

    socket
    |> assign(:page_title, "Edit Todo")
    |> assign(:todo, todo)
    |> assign(:form, to_form(Todos.change_todo(todo)))
  end

  defp apply_action(socket, :new, _params) do
    todo = %Todo{}

    socket
    |> assign(:page_title, "New Todo")
    |> assign(:todo, todo)
    |> assign(:form, to_form(Todos.change_todo(todo)))
  end

  @impl true
  def handle_event("validate", %{"todo" => todo_params}, socket) do
    changeset = Todos.change_todo(socket.assigns.todo, todo_params)
    {:noreply, assign(socket, form: to_form(changeset, action: :validate))}
  end

  def handle_event("save", %{"todo" => todo_params}, socket) do
    save_todo(socket, socket.assigns.live_action, todo_params)
  end

  defp save_todo(socket, :edit, todo_params) do
    case Todos.update_todo(socket.assigns.todo, todo_params) do
      {:ok, todo} ->
        {:noreply,
         socket
         |> put_flash(:info, "Todo updated successfully")
         |> push_navigate(to: return_path(socket.assigns.return_to, todo))}

      {:error, %Ecto.Changeset{} = changeset} ->
        {:noreply, assign(socket, form: to_form(changeset))}
    end
  end

  defp save_todo(socket, :new, todo_params) do
    case Todos.create_todo(todo_params) do
      {:ok, todo} ->
        {:noreply,
         socket
         |> put_flash(:info, "Todo created successfully")
         |> push_navigate(to: return_path(socket.assigns.return_to, todo))}

      {:error, %Ecto.Changeset{} = changeset} ->
        {:noreply, assign(socket, form: to_form(changeset))}
    end
  end

  defp return_path("index", _todo), do: ~p"/"
  defp return_path("show", todo), do: ~p"/todos/#{todo}"
end
