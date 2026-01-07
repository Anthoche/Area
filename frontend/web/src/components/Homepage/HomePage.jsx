import Navbar from "../Navbar.jsx";
import Footer from "../Footer.jsx";
import FilterTag from "./FilterTag.jsx";
import KonectCard from "./KonectCard.jsx";
import "./homepage.css";

export default function HomePage() {
    return (
        <div className="home-page-wrapper page-wrapper">
            <Navbar/>
            <div className="home-page-content">
                <div className="home-page-header home-page-section">
                    <div className="home-page-header-text">
                        <h1>Konects</h1>
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
                    <span>Search bar here</span>
                </div>
                <div className="konects-filter home-page-section">
                    <div className="filter-section">
                        <h3 className="filter-title">Type</h3>
                        <ul className="filter-buttons">
                            {/*TODO:  Replace hard-coded tags with typeFiltersList*/}
                            <li>
                                <FilterTag
                                    key={"test-filter-tag"}
                                    label={"Bonjour"}
                                    selected={true}
                                    onClick={() => console.log("clicked")}
                                />
                            </li>
                            <li>
                                <FilterTag
                                    key={"test-filter-tag-1"}
                                    label={"Bonjour 2"}
                                    selected={false}
                                    onClick={() => console.log("clicked")}
                                />
                            </li>
                            <li>
                                <FilterTag
                                    key={"test-filter-tag-2"}
                                    label={"Bonjour 3"}
                                    selected={false}
                                    onClick={() => console.log("clicked")}
                                />
                            </li>
                        </ul>
                    </div>
                    <div className="filter-section">
                        <h3 className="filter-title">Services</h3>
                        <ul className="filter-buttons">
                            {/*TODO:  Replace hard-coded tags with typeFiltersList*/}
                            <li>
                                <FilterTag
                                    key={"test-filter-tag-s1"}
                                    label={"Service 1"}
                                    selected={true}
                                    onClick={() => console.log("clicked")}
                                />
                            </li>
                            <li>
                                <FilterTag
                                    key={"test-filter-tag-s2"}
                                    label={"Service 2"}
                                    selected={false}
                                    onClick={() => console.log("clicked")}
                                />
                            </li>
                            <li>
                                <FilterTag
                                    key={"test-filter-tag-s3"}
                                    label={"Service 3"}
                                    selected={false}
                                    onClick={() => console.log("clicked")}
                                />
                            </li>
                        </ul>
                    </div>
                </div>
                <div className="konects">
                    <h2>My Konects</h2>
                    <ul className="konects-list">
                        <li>
                            <KonectCard
                                title={"Allo bonjour ici jean-damien"}
                                desc={"bonjour ceci est une description"}
                                type={"manual"}
                                services={["GitHub", "Gmail"]}
                                isActive={true}
                            />
                        </li>
                        <li>
                            <KonectCard
                                title={"Bonjour Kikonect !"}
                                desc={"bonjour ceci est une description"}
                                type={"manual"}
                                services={["Weather", "Gmail"]}
                                isActive={false}
                            />
                        </li>
                        <li>
                            <KonectCard
                                title={"Ahhh konect card"}
                                desc={"lkjdfhslkjfhdskjfhds kljfh dslkjh eslkjhdkjh"}
                                type={"timer"}
                                services={["Discord", "Dropbox"]}
                                isActive={true}
                            />
                        </li>
                    </ul>
                </div>
            </div>
            <Footer/>
            <button className="refresh-button " title="Refresh konects">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                     strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="lucide lucide-rotate-cw">
                    <path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8"></path>
                    <path d="M21 3v5h-5"></path>
                </svg>
            </button>
        </div>
    )
}