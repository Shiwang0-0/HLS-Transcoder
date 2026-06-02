const Button = ({ btnName, onClick, choice, disabled }) => {
  const isChoose = choice === 'choose'

  return (
    <button
      className={`
        flex items-center gap-2 font-semibold text-base px-6 py-3.5 text-white border-none rounded-full 
        transition-all duration-250 ease-in-out tracking-wide
        ${disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer opacity-100 hover:-translate-y-0.5 hover:scale-[1.02] active:scale-[0.98]'}
        ${isChoose 
          ? 'bg-gradient-to-br from-[#4f46e5] to-[#7c3aed] shadow-[0_10px_25px_rgba(99,102,241,0.35)]' 
          : 'bg-gradient-to-br from-[#16a34a] to-[#22c55e] shadow-[0_10px_25px_rgba(34,197,94,0.35)]'
        }
      `}
      onClick={onClick}
      disabled={disabled}
    >
      {isChoose ? (
        <svg height={24} width={24} viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" className="shrink-0">
          <path d="M0 0h24v24H0z" fill="none" />
          <path d="M11 11V5h2v6h6v2h-6v6h-2v-6H5v-2z" fill="currentColor" />
        </svg>
      ) : (
        <svg height={24} width={24} viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" className="shrink-0">
          <path d="M2 21l21-9L2 3v7l15 2-15 2z" fill="currentColor" />
        </svg>
      )}
      <span>{btnName}</span>
    </button>
  )
}

export default Button