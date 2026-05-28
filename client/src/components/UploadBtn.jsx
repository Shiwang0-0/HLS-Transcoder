import styled from 'styled-components'

const Button = ({ btnName, onClick, choice, disabled }) => {
  return (
    <StyledWrapper $choice={choice} $disabled={disabled}>
      <button className="custom-button" onClick={onClick} disabled={disabled}>
        {choice === 'choose' ? (
          <svg height={24} width={24} viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path d="M0 0h24v24H0z" fill="none" />
            <path d="M11 11V5h2v6h6v2h-6v6h-2v-6H5v-2z" fill="currentColor" />
          </svg>
        ) : (
          <svg height={24} width={24} viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path d="M2 21l21-9L2 3v7l15 2-15 2z" fill="currentColor" />
          </svg>
        )}
        <span>{btnName}</span>
      </button>
    </StyledWrapper>
  )
}

const StyledWrapper = styled.div`
  .custom-button {
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: inherit;
    cursor: ${({ $disabled }) => ($disabled ? 'not-allowed' : 'pointer')};
    font-weight: 600;
    font-size: 16px;
    padding: 0.9em 1.6em;
    color: white;
    border: none;
    border-radius: 999px;
    transition: 0.25s ease;
    letter-spacing: 0.04em;
    opacity: ${({ $disabled }) => ($disabled ? 0.6 : 1)};

    background: ${({ $choice }) =>
      $choice === 'choose'
        ? 'linear-gradient(135deg, #4f46e5, #7c3aed)'
        : 'linear-gradient(135deg, #16a34a, #22c55e)'};

    box-shadow: ${({ $choice }) =>
      $choice === 'choose'
        ? '0 10px 25px rgba(99,102,241,0.35)'
        : '0 10px 25px rgba(34,197,94,0.35)'};
  }

  .custom-button:hover:not(:disabled) {
    transform: translateY(-2px) scale(1.02);
  }

  .custom-button:active:not(:disabled) {
    transform: scale(0.98);
  }

  svg {
    flex-shrink: 0;
  }
`

export default Button