import React from 'react'

function MainContent(): React.JSX.Element {
    return <div>Vote</div>
}

function Vote(): React.JSX.Element {
    return (
        <div>
            <MainContent />
        </div>
    )
}

export default Vote
