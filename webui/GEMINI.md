# Shipyard Web UI

This directory contains the frontend application for the `shipyard` CLI tool. It provides a modern web interface for managing applications, hosts, and deployments.

## Project Overview

- **Framework:** [SolidJS](https://www.solidjs.com/) (Reactive, fine-grained updates)
- **Build Tool:** [Vite](https://vitejs.dev/)
- **Styling:** [Tailwind CSS v4](https://tailwindcss.com/) + [DaisyUI v5](https://daisyui.com/)
- **Routing:** Custom file-based routing (inspired by Next.js/SolidStart)
- **State Management:** `@tanstack/solid-query` (for server state) + `solid-js` signals/stores
- **Language:** TypeScript

## Architecture & Key Concepts

### 1. File-Based Routing
The project uses a custom file-based router located in `@router` (`src/router/FileRouter.tsx`).
- **Routes:** Defined by the folder structure in `src/routes`.
- **Conventions:**
  - `page.tsx`: The main view for a route.
  - `layout.tsx`: Wraps child routes (supports nesting).
  - `error.tsx`: Error boundary for the route.
  - `loading.tsx`: Suspense fallback for the route.
- **Dynamic Routes:** Folders with brackets (e.g., `src/routes/admin/applications/[appuid]`) become route parameters.

### 2. API Communication
- **Client:** Configured in `src/api/client.ts`.
- **Base URL:** `/api` (Proxied to `http://127.0.0.1:15678` in development).
- **Authentication:** Uses Bearer tokens stored in `localStorage` (`shipyard_access_token`).
- **Structure:**
  - `src/api/services`: Direct Axios calls to backend endpoints.
  - `src/api/hooks`: Custom Solid hooks (often wrapping TanStack Query) for consumption in components.

### 3. Build & Integration
- **Output:** The build output is targeted to `../internal/api/webui`.
- **Embedding:** The compiled static files are embedded into the Go binary using `//go:embed`.

## Development

### Prerequisites
- [Bun](https://bun.sh/) (Preferred package manager) or Node.js

### Commands

```bash
# Install dependencies
bun install

# Start development server (Port 3000)
# Proxies /api requests to a running local backend (usually on port 15678)
bun run dev

# Build for production (Outputs to ../internal/api/webui)
bun run build
```

### Configuration
- **`vite.config.ts`**: Vite configuration, including proxy settings and path aliases.
- **`tsconfig.json`**: TypeScript configuration. **Note:** Path aliases here must match `vite.config.ts`.

## Key Directories

| Directory | Description |
| :--- | :--- |
| `src/routes` | Route definitions (Pages, Layouts). |
| `src/api` | API client, services, and hooks. |
| `src/components` | Reusable UI components. |
| `src/contexts` | Global state contexts (e.g., AuthContext). |
| `src/router` | Custom routing logic (`FileRouter`). |
| `src/types` | TypeScript type definitions (shared with backend concepts). |
| `src/i18n` | Internationalization files. |

## Conventions

- **Path Aliases:** Use `@api`, `@components`, `@router`, etc., instead of relative paths.
- **Components:** Functional components using SolidJS JSX.
- **Styling:** Utility-first CSS with Tailwind. Use DaisyUI classes for components (e.g., `btn`, `card`).

## How to Add New Admin Pages (CRUD Example)

The following steps demonstrate how to add a new resource management page (e.g., managing "Products").

### 1. Define Types (`src/types/product.ts`)
First, define the data model.
```typescript
export interface Product {
  id: string;
  name: string;
  price: number;
}
```

### 2. Create API Service (`src/api/services/productService.ts`)
Use `apiClient` to define methods for interacting with the backend.
```typescript
import apiClient from '../client';
import { Product } from '@types/product';

export const productService = {
  list: async () => {
    const { data } = await apiClient.get<Product[]>('/products');
    return data;
  },
  create: async (product: Omit<Product, 'id'>) => {
    const { data } = await apiClient.post<Product>('/products', product);
    return data;
  },
  // ... update, delete
};
```

### 3. Create Hook (`src/api/hooks/useProducts.ts`)
Wrap `@tanstack/solid-query` for use in components.
```typescript
import { createQuery, createMutation, useQueryClient } from '@tanstack/solid-query';
import { productService } from '../services/productService';

export const useProducts = () => {
  const queryClient = useQueryClient();

  const query = createQuery(() => ({
    queryKey: ['products'],
    queryFn: productService.list,
  }));

  const createMutation = createMutation(() => ({
    mutationFn: productService.create,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['products'] }),
  }));

  return { query, createMutation };
};
```

### 4. Create Page (`src/routes/admin/products/page.tsx`)
Create a new route page under the `admin` directory.
```tsx
import { Component, For, createSignal } from 'solid-js';
import { useProducts } from '@api/hooks/useProducts';

const ProductsPage: Component = () => {
  const { query, createMutation } = useProducts();
  const [name, setName] = createSignal('');

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    createMutation.mutate({ name: name(), price: 100 });
  };

  return (
    <div class="p-6">
      <h1 class="text-2xl font-bold mb-4">Products</h1>
      
      {/* List */}
      <div class="grid gap-4">
        <For each={query.data}>
          {(product) => (
            <div class="card bg-base-100 shadow">
              <div class="card-body">
                {product.name} - ${product.price}
              </div>
            </div>
          )}
        </For>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} class="mt-8">
        <input 
          type="text" 
          value={name()} 
          onInput={(e) => setName(e.currentTarget.value)}
          class="input input-bordered" 
          placeholder="Product Name"
        />
        <button 
          type="submit" 
          class="btn btn-primary ml-2"
          disabled={createMutation.isPending}
        >
          Add
        </button>
      </form>
    </div>
  );
};

export default ProductsPage;
```