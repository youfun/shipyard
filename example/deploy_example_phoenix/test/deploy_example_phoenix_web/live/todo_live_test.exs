defmodule DeployExamplePhoenixWeb.TodoLiveTest do
  use DeployExamplePhoenixWeb.ConnCase

  import Phoenix.LiveViewTest
  import DeployExamplePhoenix.TodosFixtures

  @create_attrs %{description: "some description", title: "some title", completed: true}
  @update_attrs %{description: "some updated description", title: "some updated title", completed: false}
  @invalid_attrs %{description: nil, title: nil, completed: false}
  defp create_todo(_) do
    todo = todo_fixture()

    %{todo: todo}
  end

  describe "Index" do
    setup [:create_todo]

    test "lists all todos", %{conn: conn, todo: todo} do
      {:ok, _index_live, html} = live(conn, ~p"/todos")

      assert html =~ "Listing Todos"
      assert html =~ todo.title
    end

    test "saves new todo", %{conn: conn} do
      {:ok, index_live, _html} = live(conn, ~p"/todos")

      assert {:ok, form_live, _} =
               index_live
               |> element("a", "New Todo")
               |> render_click()
               |> follow_redirect(conn, ~p"/todos/new")

      assert render(form_live) =~ "New Todo"

      assert form_live
             |> form("#todo-form", todo: @invalid_attrs)
             |> render_change() =~ "can&#39;t be blank"

      assert {:ok, index_live, _html} =
               form_live
               |> form("#todo-form", todo: @create_attrs)
               |> render_submit()
               |> follow_redirect(conn, ~p"/todos")

      html = render(index_live)
      assert html =~ "Todo created successfully"
      assert html =~ "some title"
    end

    test "updates todo in listing", %{conn: conn, todo: todo} do
      {:ok, index_live, _html} = live(conn, ~p"/todos")

      assert {:ok, form_live, _html} =
               index_live
               |> element("#todos-#{todo.id} a", "Edit")
               |> render_click()
               |> follow_redirect(conn, ~p"/todos/#{todo}/edit")

      assert render(form_live) =~ "Edit Todo"

      assert form_live
             |> form("#todo-form", todo: @invalid_attrs)
             |> render_change() =~ "can&#39;t be blank"

      assert {:ok, index_live, _html} =
               form_live
               |> form("#todo-form", todo: @update_attrs)
               |> render_submit()
               |> follow_redirect(conn, ~p"/todos")

      html = render(index_live)
      assert html =~ "Todo updated successfully"
      assert html =~ "some updated title"
    end

    test "deletes todo in listing", %{conn: conn, todo: todo} do
      {:ok, index_live, _html} = live(conn, ~p"/todos")

      assert index_live |> element("#todos-#{todo.id} a", "Delete") |> render_click()
      refute has_element?(index_live, "#todos-#{todo.id}")
    end
  end

  describe "Show" do
    setup [:create_todo]

    test "displays todo", %{conn: conn, todo: todo} do
      {:ok, _show_live, html} = live(conn, ~p"/todos/#{todo}")

      assert html =~ "Show Todo"
      assert html =~ todo.title
    end

    test "updates todo and returns to show", %{conn: conn, todo: todo} do
      {:ok, show_live, _html} = live(conn, ~p"/todos/#{todo}")

      assert {:ok, form_live, _} =
               show_live
               |> element("a", "Edit")
               |> render_click()
               |> follow_redirect(conn, ~p"/todos/#{todo}/edit?return_to=show")

      assert render(form_live) =~ "Edit Todo"

      assert form_live
             |> form("#todo-form", todo: @invalid_attrs)
             |> render_change() =~ "can&#39;t be blank"

      assert {:ok, show_live, _html} =
               form_live
               |> form("#todo-form", todo: @update_attrs)
               |> render_submit()
               |> follow_redirect(conn, ~p"/todos/#{todo}")

      html = render(show_live)
      assert html =~ "Todo updated successfully"
      assert html =~ "some updated title"
    end
  end
end
