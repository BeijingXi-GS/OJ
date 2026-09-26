import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import type { ComponentProps } from 'react'
import { cn } from '@/lib/utils'

const variants = cva(
  'inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-medium transition focus-visible:outline-2 focus-visible:outline-offset-2 disabled:opacity-50 disabled:pointer-events-none',
  {
    variants: {
      variant: {
        default: 'bg-teal-700 text-white hover:bg-teal-800',
        outline: 'border border-slate-300 bg-white hover:bg-slate-50',
      },
    },
    defaultVariants: { variant: 'default' },
  },
)

export function Button({
  className,
  variant,
  asChild = false,
  ...props
}: ComponentProps<'button'> &
  VariantProps<typeof variants> & { asChild?: boolean }) {
  const Component = asChild ? Slot : 'button'
  return (
    <Component className={cn(variants({ variant }), className)} {...props} />
  )
}
