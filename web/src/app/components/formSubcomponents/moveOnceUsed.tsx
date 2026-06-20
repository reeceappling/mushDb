// TODO: use this for things like deleting entries, weird transfers (ex: Bag/box -> plate)
export function ConfirmOrCancel({txt,onConfirm,onCancel}:{onConfirm: ()=>void, txt?: string, onCancel?:()=>void}){
    if (window.confirm(txt||"Are you sure you want to do this?")){
        onConfirm && onConfirm()
    }else{
        onCancel && onCancel()
    }
}