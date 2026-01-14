/**
 * @file PopUseCases.jsx
 * @description
 * Section listing popular automation use cases.
 *
 * Allows users to:
 *  - Discover popular automation scenarios
 *  - Get inspiration for workflow creation
 */

import { useEffect, useRef, useState } from "react";
import "./welcomepage.css";
import "./hero-animations.css";

export default function PopUseCases() {
    const sectionRef = useRef(null);
    const [visible, setVisible] = useState(false);
    
    // Observe section visibility for animations
    useEffect(() => {
        const observer = new window.IntersectionObserver(
            ([entry]) => setVisible(entry.isIntersecting),
            { threshold: 0.2 }
        );
        if (sectionRef.current) observer.observe(sectionRef.current);
        return () => observer.disconnect();
    }, []);
    return (
      <div ref={sectionRef} className={`welcome-page-section popular-use-cases hero-animate${visible ? ' visible' : ''}`} id="use-cases">
          <h2>Popular use cases</h2>
          <div className="use-cases-list">
              <ul>
                  <li className={`hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.1s' : '0s' }}>Automatically save email attachments to cloud storage</li>
                  <li className={`hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.2s' : '0s' }}>Create tasks from customer support tickets</li>
                  <li className={`hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.3s' : '0s' }}>Send notifications for important events</li>
                  <li className={`hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.4s' : '0s' }}>Post new blog content to all social media channels</li>
                  <li className={`hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.5s' : '0s' }}>Sync contacts across multiple platforms</li>
                  <li className={`hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.6s' : '0s' }}>Generate reports and send to your team</li>
              </ul>
          </div>
      </div>
    )
}