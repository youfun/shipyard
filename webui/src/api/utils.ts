import { 
  useMutation, 
  useQueryClient, 
  MutationFunction,
  queryOptions,
  QueryFunction
} from '@tanstack/solid-query';

/**
 * 创建查询选项的工具函数，提供更好的类型安全性和可重用性
 * @param key 查询键
 * @param fn 查询函数
 * @param options 额外的查询选项
 * @returns 查询选项对象
 */
export function createQueryOptions<TData>(
  key: readonly unknown[],
  fn: QueryFunction<TData, readonly unknown[]>,
  options?: {
    enabled?: boolean;
    staleTime?: number;
    retry?: boolean | number;
    refetchOnWindowFocus?: boolean;
  }
) {
  return queryOptions({
    queryKey: key,
    queryFn: fn,
    ...options,
  });
}

/**
 * 封装了自动失效逻辑的 Mutation Hook (现代版本)
 * @param fn 执行异步操作的函数 (Service 层)
 * @param invalidateKeys 操作成功后需要失效（刷新）的 Query Keys 数组，可以是函数以支持动态生成
 * @param onSuccessCallback 额外的成功回调
 * @returns Mutation 对象
 */
export function useInvalidateMutation<TData, TVariables>(
  fn: MutationFunction<TData, TVariables>,
  invalidateKeys: unknown[][] | ((data: TData, variables: TVariables) => unknown[][]) = [],
  onSuccessCallback?: (data: TData, variables: TVariables) => void
) {
  const queryClient = useQueryClient();

  return useMutation(() => ({
    mutationFn: fn,
    onSuccess: (data, variables) => {
      // 支持动态和静态的invalidation keys
      const keysToInvalidate = typeof invalidateKeys === 'function' 
        ? invalidateKeys(data, variables) 
        : invalidateKeys;
      
      // 批量失效多个查询键
      keysToInvalidate.forEach(key => {
        queryClient.invalidateQueries({ queryKey: key });
      });
      
      // 执行额外的成功回调
      onSuccessCallback?.(data, variables);
    }
  }));
}

/**
 * 创建简单的 mutation，只需要函数即可
 * @param fn mutation 函数
 * @returns Mutation 对象
 */
export function useSimpleMutation<TData, TVariables>(
  fn: MutationFunction<TData, TVariables>
) {
  return useMutation(() => ({
    mutationFn: fn,
  }));
}
