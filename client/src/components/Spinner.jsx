const Spinner = () => (
  <div
    style={{
      width: '18px',
      height: '18px',
      border: '2px solid #e5e7eb',
      borderTop: '2px solid #4f46e5',
      borderRadius: '50%',
      animation: 'spin 0.8s linear infinite',
      flexShrink: 0,
    }}
  >
    <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>
  </div>
)

export default Spinner