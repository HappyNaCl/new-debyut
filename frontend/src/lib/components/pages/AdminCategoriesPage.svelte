<script lang="ts">
  import "../../../styles/admin.sass";
  import { Add, Delete, Edit } from "@material-symbols-svg/svelte";
  import AdminLayout from "../admin/AdminLayout.svelte";
  import CategoryForm from "../admin/CategoryForm.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import {
    categoriesQueryKey,
    createCategoriesQuery,
    createCategory,
    deleteCategory,
    updateCategory,
  } from "../../api/categories";
  import { getCategoryIcon } from "../../icons/categoryIcons";
  import { queryClient } from "../../queryClient";
  import type { CategoryInput, EventCategory } from "../../types/category";

  type Editor = { mode: "create" } | { mode: "edit"; category: EventCategory } | null;

  const categories = createCategoriesQuery();

  let editor = $state<Editor>(null);
  let confirmingDeleteId = $state<string | null>(null);
  let deletingId = $state<string | null>(null);
  let errorMessage = $state("");
  let successMessage = $state("");

  function refresh() {
    return queryClient.invalidateQueries({ queryKey: categoriesQueryKey });
  }

  function openEditor(next: Editor) {
    editor = next;
    confirmingDeleteId = null;
    errorMessage = "";
    successMessage = "";
  }

  async function handleSubmit(input: CategoryInput) {
    if (editor?.mode === "edit") {
      const updated = await updateCategory(editor.category.id, input);
      successMessage = `Saved ${updated.name}.`;
    } else {
      const created = await createCategory(input);
      successMessage = `Created ${created.name}.`;
    }
    editor = null;
    await refresh();
  }

  async function handleDelete(category: EventCategory) {
    errorMessage = "";
    successMessage = "";
    deletingId = category.id;
    try {
      await deleteCategory(category.id);
      successMessage = `Deleted ${category.name}.`;
      if (editor?.mode === "edit" && editor.category.id === category.id) editor = null;
      await refresh();
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      deletingId = null;
      confirmingDeleteId = null;
    }
  }
</script>

<div class="admin-root">
  <AdminLayout title="Event categories">
    <div class="page-heading">
      <div>
        <p>Categories appear in the header menu and on event pages.</p>
      </div>
      <button type="button" class="button-primary" onclick={() => openEditor({ mode: "create" })}>
        <Add size={20} /> New category
      </button>
    </div>

    {#if successMessage}
      <div class="alert alert-success" role="status">{successMessage}</div>
    {/if}
    {#if errorMessage}
      <div class="alert alert-error" role="alert">{errorMessage}</div>
    {/if}

    {#if editor}
      <section class="card">
        {#key editor}
          <CategoryForm
            category={editor.mode === "edit" ? editor.category : undefined}
            onSubmit={handleSubmit}
            onCancel={() => (editor = null)}
          />
        {/key}
      </section>
    {/if}

    {#if categories.isPending}
      <div class="state"><Spinner size={32} /></div>
    {:else if categories.isError}
      <div class="state">
        <p>{extractErrorMessage(categories.error)}</p>
        <button type="button" class="button-outline" onclick={() => categories.refetch()}>Try again</button>
      </div>
    {:else if categories.data.length === 0}
      <div class="state">
        <p>No categories yet. Create the first one.</p>
      </div>
    {:else}
      <ul class="item-list">
        {#each categories.data as category (category.id)}
          {@const Icon = getCategoryIcon(category.icon)}
          <li class="item-row">
            <span class="item-icon"><Icon size={22} /></span>
            <div class="item-text">
              <span class="item-title">{category.name}</span>
              <span class="item-detail">{category.slug}</span>
            </div>

            {#if confirmingDeleteId === category.id}
              <div class="row-actions">
                <span class="confirm-text">Delete?</span>
                <button
                  type="button"
                  class="button-outline"
                  onclick={() => (confirmingDeleteId = null)}
                  disabled={deletingId === category.id}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  class="button-danger"
                  onclick={() => handleDelete(category)}
                  disabled={deletingId === category.id}
                >
                  {deletingId === category.id ? "Deleting…" : "Delete"}
                </button>
              </div>
            {:else}
              <div class="row-actions">
                <button
                  type="button"
                  class="icon-button"
                  aria-label={`Edit ${category.name}`}
                  title="Edit"
                  onclick={() => openEditor({ mode: "edit", category })}
                >
                  <Edit size={20} />
                </button>
                <button
                  type="button"
                  class="icon-button danger"
                  aria-label={`Delete ${category.name}`}
                  title="Delete"
                  onclick={() => (confirmingDeleteId = category.id)}
                >
                  <Delete size={20} />
                </button>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </AdminLayout>
</div>
