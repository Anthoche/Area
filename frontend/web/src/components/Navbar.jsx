import logo from "../../lib/assets/Kikonect_logo_no_text.png";
import React, { useEffect, useState } from "react";

const API_BASE =
    import.meta.env.VITE_API_URL ||
    import.meta.env.API_URL ||
    `${window.location.protocol}//${window.location.hostname}:8080`;

export default function Navbar() {
    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const readAuthInfo = () => ({
        userId: Number(localStorage.getItem("user_id")),
        userEmail: localStorage.getItem("user_email") || "",
        googleTokenId: localStorage.getItem("google_token_id"),
        githubTokenId: localStorage.getItem("github_token_id"),
    });
    const [authInfo, setAuthInfo] = useState(readAuthInfo);

    const toggleMenu = () => {
        setIsMenuOpen(!isMenuOpen);
    };
    const closeMenu = () => {
        setIsMenuOpen(false);
    };

    useEffect(() => {
        const refresh = () => setAuthInfo(readAuthInfo());
        window.addEventListener("storage", refresh);
        window.addEventListener("auth-updated", refresh);
        return () => {
            window.removeEventListener("storage", refresh);
            window.removeEventListener("auth-updated", refresh);
        };
    }, []);

    const userId = authInfo.userId;
    const userEmail = authInfo.userEmail;
    const googleTokenId = authInfo.googleTokenId;
    const githubTokenId = authInfo.githubTokenId;
    const isLoggedIn = Number.isFinite(userId) && userId > 0;
    const logout = () => {
        localStorage.clear();
        window.dispatchEvent(new Event("auth-updated"));
        window.location.href = "/";
    };
    const oauthRedirect = encodeURIComponent(`${window.location.origin}/home`);
    const connectGoogle = () => {
        const baseUrl = `${API_BASE}/oauth/google/login?ui_redirect=${oauthRedirect}`;
        const url = userId && userId > 0 ? `${baseUrl}&user_id=${userId}` : baseUrl;
        window.location.href = url;
    };
    const connectGithub = () => {
        const baseUrl = `${API_BASE}/oauth/github/login?ui_redirect=${oauthRedirect}`;
        const url = userId && userId > 0 ? `${baseUrl}&user_id=${userId}` : baseUrl;
        window.location.href = url;
    };

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
                        {isLoggedIn && userEmail && (
                            <li className="navbar-email">{userEmail}</li>
                        )}
                        {isLoggedIn && !googleTokenId && (
                            <>
                                <li className="navbar-connect-btn">
                                    <button type="button" onClick={connectGoogle}>
                                        Connect Google
                                    </button>
                                </li>
                            </>
                        )}
                        {isLoggedIn && !githubTokenId && (
                            <li className="navbar-connect-btn">
                                <button type="button" onClick={connectGithub}>
                                    Connect GitHub
                                </button>
                            </li>
                        )}
                        {isLoggedIn ? (
                            <li className="navbar-logout-btn">
                                <button type="button" onClick={logout}>
                                    Logout
                                </button>
                            </li>
                        ) : (
                            <li className="navbar-login-btn">
                                <a href="/login" onClick={closeMenu}>
                                    Login
                                </a>
                            </li>
                        )}
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
                        {isLoggedIn && userEmail && (
                            <li className="navbar-email">{userEmail}</li>
                        )}
                        {isLoggedIn && !googleTokenId && (
                            <>
                                <li className="navbar-connect-btn">
                                    <button type="button" onClick={connectGoogle}>
                                        Connect Google
                                    </button>
                                </li>
                            </>
                        )}
                        {isLoggedIn && !githubTokenId && (
                            <li className="navbar-connect-btn">
                                <button type="button" onClick={connectGithub}>
                                    Connect GitHub
                                </button>
                            </li>
                        )}
                        {isLoggedIn ? (
                            <li className="navbar-logout-btn">
                                <button type="button" onClick={logout}>
                                    Logout
                                </button>
                            </li>
                        ) : (
                            <li className="mobile-login-btn">
                                <a href="/login" onClick={closeMenu}>
                                    Login
                                </a>
                            </li>
                        )}
                    </ul>
                </div>
            </div>
        </div>
    )
}
