
function Navbar() {
    return <nav>
        <h1>Bob's Bistro</h1>
        <ul>
            <li>Menu</li>4
            <li>About</li>
            <li>Contact</li>
        </ul>
    </nav>
}

function MainContent() {
    return (
        <div>
            <h1 className="header">This is JSX</h1>
            <p>This is a paragraph</p>
        </div>
    )
}

function Index() {
    return (
        <div>
            <Navbar/>
            <MainContent/>
        </div>
    )
}

export default Index;
