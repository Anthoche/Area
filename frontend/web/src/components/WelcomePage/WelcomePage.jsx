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
                    <div className="welcome-page-section-pres">
                        <div className="welcome-page-section-pres-icon">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                                 stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                 className="lucide lucide-sparkles" aria-hidden="true">
                                <path
                                    d="M11.017 2.814a1 1 0 0 1 1.966 0l1.051 5.558a2 2 0 0 0 1.594 1.594l5.558 1.051a1 1 0 0 1 0 1.966l-5.558 1.051a2 2 0 0 0-1.594 1.594l-1.051 5.558a1 1 0 0 1-1.966 0l-1.051-5.558a2 2 0 0 0-1.594-1.594l-5.558-1.051a1 1 0 0 1 0-1.966l5.558-1.051a2 2 0 0 0 1.594-1.594z"></path>
                                <path d="M20 2v4"></path>
                                <path d="M22 4h-4"></path>
                                <circle cx="4" cy="20" r="2"></circle>
                            </svg>
                            <span>Connect your apps and automate workflows</span>
                        </div>
                        <h1>Make your apps work together seamlessly</h1>
                        <p>
                            Connect your favorite apps and services to automate workflows. No code required.
                            Create powerful automations in minutes and boost your productivity.
                        </p>
                        <div className="welcome-page-section-pres-btns">
                            <div className="pres-btn-get-started pres-btn">
                                <span>Get Started</span>
                                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                                     stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                                     className="lucide lucide-arrow-right group-hover:translate-x-1 transition-transform"
                                     aria-hidden="true">
                                    <path d="M5 12h14"></path>
                                    <path d="m12 5 7 7-7 7"></path>
                                </svg>
                            </div>
                            <div className="pres-btn pres-btn-demo">
                                <span>Watch demo</span>
                            </div>
                        </div>
                    </div>
                    <div className="welcome-page-section-stats">
                        <ul>
                            <li>
                                <h1>10+</h1>
                                <span>Active users</span>
                            </li>
                            <li>
                                <h1>10+</h1>
                                <span>Available services</span>
                            </li>
                            <li>
                                <h1>50+</h1>
                                <span>Actions/Reactions</span>
                            </li>
                        </ul>
                    </div>
                </div>
                <div className="welcome-page-section" id="how-it-works">
                    how it works
                </div>
            </div>
            <Footer/>
        </div>
    )
}