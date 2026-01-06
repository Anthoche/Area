import logo from "../../../lib/assets/Kikonect_logo_no_text.png";
import React, {useState} from "react";
import { Link } from "react-router-dom";

export default function Navbar() {
    const [isMenuOpen, setIsMenuOpen] = useState(false);

    const toggleMenu = () => {
        setIsMenuOpen(!isMenuOpen);
    };

    const closeMenu = () => {
        setIsMenuOpen(false);
    };

    const userId = Number(localStorage.getItem("user_id"));
    const isLoggedIn = Number.isFinite(userId) && userId > 0;
    

    return (
        <div className="navbar-wrapper">
            <nav>
                <div className="navbar-section">
                    <ul className="navbar-list navbar-list-logo">
                        <li><img src={logo} alt="KiKoNect logo" className="logo-navbar"/></li>
                        <li><h2>KiKoNect</h2></li>
                    </ul>
                </div>
                <div className="navbar-section">
                    <ul className="navbar-list">
                        <li><a href="#features">Features</a></li>
                        <li><a href="#how-it-works">How it works</a></li>
                        <li className="navbar-login-btn">
                            <a 
                                href={isLoggedIn ? "/home" : "/login"} 
                                onClick={closeMenu}
                            >
                            {isLoggedIn ? "My Account" : "Login"}
                            </a>
                        </li>
                    </ul>
                </div>
                <button className="hamburger-button" onClick={toggleMenu} aria-label="Toggle menu">
                    <span className={`hamburger-line ${isMenuOpen ? 'open' : ''}`}></span>
                    <span className={`hamburger-line ${isMenuOpen ? 'open' : ''}`}></span>
                    <span className={`hamburger-line ${isMenuOpen ? 'open' : ''}`}></span>
                </button>
            </nav>

            <div className={`mobile-menu-overlay ${isMenuOpen ? 'open' : ''}`} onClick={closeMenu}></div>
            <div className={`mobile-menu ${isMenuOpen ? 'open' : ''}`}>
                <div className="mobile-menu-section">
                    <ul className="mobile-menu-list navbar-list-logo">
                        <li><img src={logo} alt="KiKoNect logo" className="logo-navbar"/></li>
                        <li><h2>KiKoNect</h2></li>
                    </ul>
                </div>
                <div className="mobile-menu-divider"></div>
                <div className="mobile-menu-section">
                    <ul className="mobile-menu-list">
                        <li><a href="#features" onClick={closeMenu}>Features</a></li>
                        <li><a href="#how-it-works" onClick={closeMenu}>How it works</a></li>
                        <li className="mobile-login-btn">
                            <a 
                                href={isLoggedIn ? "/home" : "/login"} 
                                onClick={closeMenu}
                            >
                            {isLoggedIn ? "My Account" : "Login"}
                            </a>
                        </li>
                    </ul>
                </div>
            </div>
        </div>
    )
}
