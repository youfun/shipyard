import { JSX } from 'solid-js';
import { Link } from '@router';

interface ButtonLinkProps {
  href: string;
  children: JSX.Element;
  class?: string;
}

export function ButtonLink(props: ButtonLinkProps) {
  return (
    <Link 
      href={props.href} 
      class={`inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white rounded-full text-sm font-medium hover:bg-indigo-700 transition-all shadow-sm hover:shadow-md ${props.class || ''}`}
    >
      {props.children}
    </Link>
  );
}
