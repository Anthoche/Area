import logo from "../../lib/assets/Kikonect_logo_no_text.png";
import React from "react";

export default function Footer() {
    return (
        <div className="footer-wrapper">
            <footer>
                <div className="footer-sections">
                    <ul className="footer-section">
                        <li className="footer-section-title footer-section-logo">
                            <img src={logo} alt="KiKoNect logo" className="logo-footer"/>
                            <span>KiKoNect</span>
                        </li>
                        <li className="footer-section-item">Making automation accessible to everyone.</li>
                    </ul>
                    <ul className="footer-section">
                        <li className="footer-section-title">Product</li>
                        <li className="footer-section-item"><a href="/#features">Features</a></li>
                        <li className="footer-section-item"><a href="/#how-it-works">How it works</a></li>
                        <li className="footer-section-item"><a href="#">Services</a></li>
                        <li className="footer-section-item"><a href="#">Enterprise</a></li>
                    </ul>
                    <ul className="footer-section">
                        <li className="footer-section-title">Resources</li>
                        <li className="footer-section-item"><a href="#">Documentation</a></li>
                        <li className="footer-section-item"><a href="#">Help Center</a></li>
                        <li className="footer-section-item"><a href="#">Community</a></li>
                        <li className="footer-section-item"><a href="#">Blog</a></li>
                    </ul>
                    <ul className="footer-section">
                        <li className="footer-section-title">Company</li>
                        <li className="footer-section-item"><a href="#">About Us</a></li>
                        <li className="footer-section-item"><a href="#">Careers</a></li>
                        <li className="footer-section-item"><a href="#">Privacy</a></li>
                        <li className="footer-section-item"><a href="#">Terms</a></li>
                    </ul>
                </div>
                <p>© 2025 KiKoNect. All rights reserved.</p>
            </footer>
        </div>
    )
}