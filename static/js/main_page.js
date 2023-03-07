/*
*****************************************************************************************
Shopping Cart
*****************************************************************************************
*/

const btn = document.querySelectorAll(".items button");
const basket = document.getElementById("basket");
const sCart = document.getElementById("liveToastBtn");
const toastBody = document.querySelector("#toastBodyTable tbody");
const arr = [];
// loop the buttons
for (let button of btn) {
  // get the button which is clicked
  button.addEventListener("click", (el) => {
    // get the elements from that button
    const id = el.target.dataset.id;
    const price = el.target.dataset.price;
    const product = el.target.dataset.product;
    
    // add the values to the array
    addItems(id, price, product);
    // display the result 
    outputItems(arr);
    
  });
}


function addItems(ids, prices, products) {
 
  // find the position of that clicked item in array
  const find = arr.map((item) => item.id).indexOf(ids);
  // if its not exisst add one to array
  if (find == -1) {
    arr.unshift({ id: ids, product: products, basPrice: prices, sumPrice: prices, quantity: 1 });
  }
  // if its exist
  if (find >= 0) {
    // get the quantity of that item from array and add 1 on it
    const newQnty = (arr[find]["quantity"] += 1);
    // splice is just doing this like, find the poition of that item in array and remove it and replace with
    // new item with increased quantity
    arr.splice(find, 1, {
      id: ids,
      product: products,
      basPrice: prices,
      sumPrice: prices * newQnty,
      quantity: newQnty
    });
  }
}






// increase item from the array
// function calls from the shopping cart output
function increaseItems(target){
    //find the position
  const find = arr.map((item) => item.id).indexOf(target.id);
    // increase the quantity
  const newQnty = (arr[find]["quantity"] += 1);
  const basPrice = (arr[find]["basPrice"]);
  const products = (arr[find]["product"]);
  
  arr.splice(find, 1, {
      id: target.id,
      product: products,
      basPrice: basPrice,
      sumPrice: basPrice * newQnty,
      quantity: newQnty
    });
  
  // output the result
  outputItems(arr);
}


// decrease item from the array
// function calls from the shopping cart output

function decreaseItems(target) {
  //find the position
  const idx = arr.map((item) => item.id).indexOf(target.id);
  // decrease the quantity
  const newQnty = (arr[idx]["quantity"] -= 1);
  // remove and replace like we did with adding
  arr.splice(idx, 1, {
    id: target.id,
    product: arr[idx].product,
    basPrice: target.dataset.basprice,
    sumPrice: target.dataset.basprice * newQnty,
    quantity: newQnty
  });
  // if quantity is zero then make price and quantity zero
  if (newQnty <= 0) {
    arr[idx].product = arr[idx].product;
    arr[idx].sumPrice = 0;
    arr[idx].quantity = 0;
   delete arr[idx];
    console.log(arr);

  }  

 // output the result
  outputItems(arr);
}



function outputItems(items) {
	let html = "";
    let goods = "?cart=";
  
  if( Object.keys(items).length < 1){ 
    html += "<th colspan='4' class='text-start'>Shopping Cart is empty</th>";
   }else{
    items.map((e, i) => {
      html +=
        "<tr><th class='text-start'>" +
        e.product +
        "</th><th>" +
        e.quantity +
        "</th><th>" +
        e.sumPrice +
        "</th><th class='ps-5'> <button class='px-2' data-unitPrice='" +
        e.unitPrice +
        "' id='" +
        e.id +
        "' onclick='decreaseItems(this)'> - </button> <button data-unitPrice='" +
        e.unitPrice +
        "' id='" +
        e.id +
        "' onclick='increaseItems(this)'>+</button></th></tr><br>";
      goods += "|" + e.quantity + ":" + e.id;
    });
      html += "<a href='"+ goods +"' class='text-white'>Checkout</a>";
   }

  
    // with every add to basket show toast on side
    const toastTrigger = document.getElementById("liveToastBtn");
    const toastLiveExample = document.getElementById("liveToast");
    const toast = new bootstrap.Toast(toastLiveExample);
    toastBody.innerHTML = html;
    toast.show();

}


 // open and show inside the shoppingcart with click
  const toastLiveExample = document.getElementById("liveToast");
  const toast = new bootstrap.Toast(toastLiveExample);
  sCart.addEventListener("click", () => { toast.show(); });