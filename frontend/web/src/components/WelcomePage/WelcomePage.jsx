"use client"

import "./welcomepage.css";
import Navbar from "./Navbar.jsx";
import Footer from "./Footer.jsx";

export default function WelcomePage() {
    return (
        <div className="welcome-page-wrapper">
            <Navbar/>
            <div className="welcome-page-content">
                <div className="welcome-page-section" id="features">
                    features
                </div>
                <div className="welcome-page-section" id="how-it-works">
                    how it works
                </div>
            </div>
            <Footer/>
        </div>
    )
}