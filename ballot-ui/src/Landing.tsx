import './Landing.css'


function MainContent() {
    return (
        <div id="Landing">
            <div className="brand">
                <span>Ballot</span>
            </div>
            <div className="form">
                <form>
                    <label htmlFor=""></label>
                    <input type="text"
                           maxLength={64}
                           placeholder="Your name/alias"/>
                    <button type="button">New Voting Space
                    </button>
                </form>
            </div>
            <div className="footer">
                <div className="renegade">Renegade Otter</div>
                <div className="sauce">Open Sauce</div>
            </div>
        </div>
    )
}

function Landing() {
    return (
        <div>
            <MainContent/>
        </div>
    )
}

export default Landing;
