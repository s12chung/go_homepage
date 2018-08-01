module.exports = function(source) {
    // this is used in the source param
    let __webpack_public_path__ = "";

    let result = "";
    eval(source.replace((/module\.exports\s?=/), "result ="));
    return JSON.stringify(result, null, 2);
};