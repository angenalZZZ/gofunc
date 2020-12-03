function test() {
    console.log('hello js');
    return 'hello golang';
}

// module.exports = {
//     test: test,
// };
exports = {
    test: test,
};
