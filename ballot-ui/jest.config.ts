export {}
/** @type {import('ts-jest/dist/types').InitialOptionsTsJest} */
module.exports = {
    preset: 'ts-jest',
    transform: {
        '^.+\\.tsx?$': 'ts-jest',
    },
    moduleNameMapper: {
        // Force module uuid to resolve with the CJS entry point, because Jest does not support package.json.exports.
        // See https://github.com/uuidjs/uuid/issues/451
        uuid: require.resolve('uuid'),
        // if your using tsconfig.paths thers is no harm in telling jest
        '@components/(.*)$': '<rootDir>/src/components/$1',
        '@/(.*)$': '<rootDir>/src/$1',

        // mocking assets and styling
        '^.+\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga)$':
            '<rootDir>/tests/mocks/fileMock.ts',
        '^.+\\.(css|less|scss|sass)$': '<rootDir>/tests/mocks/styleMock.ts',
        /* mock models and services folder */
        '(assets|models|services)': '<rootDir>/tests/mocks/fileMock.ts',
    },
    // to obtain access to the matchers.
    setupFilesAfterEnv: ['./tests/setupTests.ts'],

    moduleFileExtensions: ['ts', 'tsx', 'js', 'jsx', 'json', 'node'],
    modulePaths: ['<rootDir>'],
    testEnvironment: 'jest-fixed-jsdom',

    // https://github.com/mswjs/msw/issues/1786
    testEnvironmentOptions: {
        customExportConditions: [''],
    },
}
