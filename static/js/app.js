function Prompt(){
			let toast=function(c){
				const {
					icon= "success",
					msg="",
					position="top-end",
				}=c;
				const Toast = Swal.mixin({
				  toast: true,
				  position: position,
				  showConfirmButton: false,
				  icon: icon,
				  title: msg,
				  timer: 3000,
				  timerProgressBar: true,
				  didOpen: (toast) => {
				    toast.addEventListener('mouseenter', Swal.stopTimer)
				    toast.addEventListener('mouseleave', Swal.resumeTimer)
				  }
				});

				Toast.fire({
				 
				});
			}
			let success=function(c){
				const{
					title="Success",
					text="",
				}=c;
				Swal.fire({
				  icon: 'success',
				  title: title,
				  text: text,
				});

			}
			let error=function(c){
				const{
					title="",
					text="",
				}=c;
				Swal.fire({
				  icon: 'error',
				  title: title,
				  text: text,
				});

			}
			async function custom(c){
				const{
					msg="",
					title="",
					icon = "",
					showConfirmButton=true,
				}=c;
				const { value: result } = await Swal.fire({
				  title: title,
				  html:msg,
				  icon: icon,
				  backdrop: false,
				  showCancelButton:true,
				  focusConfirm: false,
				  showConfirmButton,
				  willOpen: () =>{
					if(c.willOpen!==undefined){
						c.willOpen();
					}
				  },
				  didOpen: () =>{
					if(c.didOpen!==undefined){
						c.didOpen();
					}
				  },
				})
				if(result.dismiss!==Swal.DismissReason.cancel){
					if(result.value!==""){
						if(c.callback!==undefined){
							c.callback(result);
						}else{
							c.callback(false);
						}
					}
				}
			}

			return{
				success : success,
				toast : toast,
				error : error,
				custom : custom,
			}
			

if (formValues) {
  Swal.fire(JSON.stringify(formValues))
}
		}