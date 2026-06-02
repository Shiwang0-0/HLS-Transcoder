import "../helpers/time"
import { timeAgo } from "../helpers/time"
import { StatusPill } from "./StatusPill"

export function VideoCard({ video, isActive, onClick }) {
  const title = (video.videoName || "Untitled").replace(/\.(mp4|mov|avi|mkv)$/i, "")
  const canPlay = video.status === "completed"

  return (
    <div
      onClick={() => canPlay && onClick(video)}
      className={`
        rounded-8 overflow-hidden transition-all duration-150 rounded-lg border
        ${isActive ? "bg-[#1a1a1a] border-[#3b82f6] -translate-y-[1px]" : "bg-[#111] border-[#222]"}
        ${canPlay ? "cursor-pointer" : "cursor-default"}
        ${video.status === "failed" ? "opacity-60" : "opacity-100"}
        ${!isActive && canPlay ? "hover:border-[#333] hover:-translate-y-0.5" : ""}
      `}
    >
      <div className="w-full aspect-video bg-[#0a0a0a] flex items-center justify-center relative overflow-hidden">
        <div 
          className="absolute inset-0 opacity-60" 
          style={{ backgroundImage: "linear-gradient(#1a1a1a 1px, transparent 1px), linear-gradient(90deg, #1a1a1a 1px, transparent 1px)", backgroundSize: "20px 20px" }} 
        />
        
        {canPlay ? (
          <div className={`relative z-1 w-9 h-9 rounded-full flex items-center justify-center border transition-colors duration-150 ${isActive ? "bg-[#3b82f6] border-[#3b82f6]" : "bg-[#1e1e1e] border-[#333]"}`}>
            <svg width="14" height="14" viewBox="0 0 24 24" fill={isActive ? "#fff" : "#888"}>
              <polygon points="5,3 19,12 5,21" />
            </svg>
          </div>
        ) : (
          <div className="relative z-1 text-[#333] text-2xl">
            {video.status === "failed" ? "✕" : "⋯"}
          </div>
        )}
        
        {isActive && (
          <div className="absolute bottom-1.5 right-1.5 z-2 bg-[#3b82f6] rounded-[3px] px-1.5 py-0.5 text-[9px] text-white font-mono tracking-wider uppercase">
            playing
          </div>
        )}
      </div>

      <div style={{ padding: "16px 18px 20px 18px" }}>
        <p className="text-[13px] font-medium text-[#e5e5e5] leading-snug mb-3 line-clamp-2 ">
          {title}
        </p>
        
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginTop: "14px"}}>
          <StatusPill status={video.status} />
          <span className="text-[11px] text-[#444] font-mono">
            {timeAgo(video.createdAt)}
          </span>
        </div>
      </div>
    </div>
  )
}