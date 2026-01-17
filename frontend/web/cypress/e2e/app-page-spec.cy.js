import {URL, USER_MAIL, USER_PASSWORD} from "./vars.js";

describe('App tests', () => {
    it('Login', () => {
        cy.visit(URL + "/login")
        cy.get(':nth-child(1) > input').type(USER_MAIL)
        cy.get(':nth-child(2) > input').type(USER_PASSWORD)
        cy.get('[type="submit"]').click()
        cy.url().should('eql', URL + '/home')
        cy.get('.navbar-list > .navbar-email').contains(USER_MAIL)
    })

    it('Logout', () => {
        cy.visit(URL + "/login")
        cy.get(':nth-child(1) > input').type(USER_MAIL)
        cy.get(':nth-child(2) > input').type(USER_PASSWORD)
        cy.get('[type="submit"]').click()
        cy.url().should('eql', URL + '/home')

        cy.visit(URL + "/home")
        cy.get('.navbar-list > .navbar-logout-btn > button').click()
        cy.url().should('eql', URL + '/')
    })
});

describe('Action flows', () => {
    beforeEach(() => {
        cy.visit(URL + "/login")
        cy.get(':nth-child(1) > input').type(USER_MAIL)
        cy.get(':nth-child(2) > input').type(USER_PASSWORD)
        cy.get('[type="submit"]').click()
        cy.url().should('eql', URL + '/home')
    })

    afterEach(() => {
        cy.get('.navbar-list > .navbar-logout-btn > button').click()
        cy.url().should('eql', URL + '/')
    })

    it('Create konect', () => {
        cy.get('.create-konect-btn').click()
        cy.get('.panel-body input').first().clear().type('Cypress Test Konect')

        cy.get('select').first().select(0)
        cy.get('select').last().select(0)
        cy.get('.primary-btn').contains('Create Konect').click()
        cy.get('.konects-list').should('contain', 'Cypress Test Konect')
    })

    it('Filtering tests', () => {
        cy.get('.show-more-button').click()

        // Filter by type
        const typesSection = cy.get('.filter-section').first();

        typesSection.find('.filter-tag').contains("Timer").click()
        typesSection.get('.filter-tag.active').should('contain', 'Timer')
        typesSection.get('.filter-tag.active').should('not.contain.value', 'All')

        // Filter by service
        const servicesSection = cy.get('.filter-section').last();

        servicesSection.find('.filter-tag').contains('Air').click()
        servicesSection.get('.filter-tag.active').should('contain', 'Air')
        servicesSection.get('.filter-tag.active').should('not.contain.value', 'All')

        // Reset filters
        cy.get('.filter-section').first().find('.filter-tag').first().click()
        cy.get('.filter-section').last().find('.filter-tag').first().click()
    })

    it('Search konect', () => {
        cy.get('.search-input').type('Cypress Test')
        cy.get('.konects-list').should('contain', 'Cypress Test')
    })

    it('Editing konect payload', () => {
        cy.get('.konect-card').first().click()
        cy.get('.payload-area').should('be.visible')
        cy.get('.payload-area').clear().type('{"test": "updated"}', {parseSpecialCharSequences: false})

        cy.get('.payload-area').should('have.value', '{"test": "updated"}')
        cy.get('.panel-close').click()
    })

    it('Delete konect', () => {
        cy.get('.konect-card').first().click()
        cy.get('.danger-btn').contains('Delete Konect').click()
    })

    it('Refresh konects', () => {
        cy.get('.refresh-button').click()
        cy.get('.refresh-button').should('not.be.disabled')
    })
})