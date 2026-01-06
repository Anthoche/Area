import Navbar from "../Navbar.jsx";
import Footer from "../Footer.jsx";
import "./homepage.css";

export default function HomePage() {
    return (
        <div className="home-page-wrapper page-wrapper">
            <Navbar/>
            <div className="home-page-content">
                <div className="home-page-header">
                    <div className="home-page-header-text">
                        <h1>My Konects</h1>
                        <span>Manage and automate your favorite services seamlessly.</span>
                        <span>Create and organize your Konects to boost productivity.</span>
                    </div>
                    <div className="home-page-create-button">
                        <button className="create-konect-btn">
                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                                 stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                 className="lucide lucide-plus"
                                 data-fg-b4eh13="107.23:107.5728:/src/app/components/KonectsContent.tsx:85:13:2780:18:e:Plus::::::MOt">
                                <path d="M5 12h14"></path>
                                <path d="M12 5v14"></path>
                            </svg>
                            <span>Create a Konect</span>
                        </button>
                    </div>
                </div>
                <div className="konects-searchbar">

                </div>
                <div className="konects-filter">
                    <div className="filter-section">
                        <h3 className="filter-title">Type</h3>
                        <ul className="filter-buttons">
                            <li>1</li>
                            <li>2</li>
                            <li>3</li>
                        </ul>
                    </div>
                    <div className="filter-section">
                        <h3 className="filter-title">Services</h3>
                        <ul className="filter-buttons">
                            <li>1</li>
                            <li>2</li>
                            <li>3</li>
                        </ul>
                    </div>
                </div>
                <div className="konects">
                    <h2>My Konects</h2>
                    <ul>
                        <li>1</li>
                        <li>2</li>
                        <li>3</li>
                    </ul>
                </div>
            </div>
            <Footer/>
        </div>
    )
}