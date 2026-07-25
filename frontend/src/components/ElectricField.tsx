export function ElectricField({ className = '' }: { className?: string }) {
  return (
    <div className={`pointer-events-none absolute inset-0 overflow-hidden ${className}`} aria-hidden>
      <div className="soft-wash absolute inset-0" />
      <div className="soft-wash-orb soft-wash-orb-a absolute -left-[20%] -top-[30%] h-[70%] w-[70%] rounded-full" />
      <div className="soft-wash-orb soft-wash-orb-b absolute -bottom-[25%] -right-[15%] h-[65%] w-[65%] rounded-full" />
      <div className="soft-wash-orb soft-wash-orb-c absolute left-[35%] top-[40%] h-[45%] w-[50%] rounded-full" />
    </div>
  )
}
