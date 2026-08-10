export function StatusPill({ status }) {
  const map = {
    completed:   { label: "ready",        bg: "bg-[#0d2e1a]", color: "text-[#4ade80]", dot: "bg-[#22c55e]" },
    transcoding: { label: "transcoding",  bg: "bg-[#2a1f00]", color: "text-[#fbbf24]", dot: "bg-[#f59e0b]" },
    downloading: { label: "downloading",  bg: "bg-[#2e1065]", color: "text-[#c4b5fd]", dot: "bg-[#8b5cf6]" },
    uploading:   { label: "uploading",  bg: "bg-[#082f49]", color: "text-[#67e8f9]", dot: "bg-[#06b6d4]" },
    pending_upload: { label: "starting",  bg: "bg-[#082f49]", color: "text-[#67e8f9]", dot: "bg-[#06b6d4]" },
    uploaded:       { label: "uploaded",  bg: "bg-[#082f49]", color: "text-[#67e8f9]", dot: "bg-[#06b6d4]" },
    queued:      { label: "queued",       bg: "bg-[#1a1a2e]", color: "text-[#818cf8]", dot: "bg-[#6366f1]" },
    failed:      { label: "failed",       bg: "bg-[#2e0d0d]", color: "text-[#f87171]", dot: "bg-[#ef4444]" },
  }
  
  const s = map[status] || map.queued
  const shouldPulse = status !== "completed" && status !== "failed"

  return (
    <span className={`inline-flex items-center gap-1.5 font-mono text-[10px] leading-none px-2.5 py-1 rounded tracking-wide uppercase ${s.bg} ${s.color}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${s.dot} ${shouldPulse ? 'animate-pulse' : ''}`} />
      {s.label}
    </span>
  )
}