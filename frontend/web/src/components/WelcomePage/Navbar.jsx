import logo from "../../../lib/assets/Kikonect_logo_no_text.png";
import React from "react";

export default function Navbar() {
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
                        <li><a href="#">Pricing</a></li>
                        <li className="navbar-login-btn"><a href="/login">Login</a></li>
                    </ul>
                </div>
            </nav>
        </div>
    )
}
