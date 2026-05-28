'use client'
export default function TestAndValidate(props: React.PropsWithChildren<{todos?:string[]}>) {
    return <>
        {props.children}
        {props.todos && props.todos.map((todo, index)=>{
            return <div key={index} className={"needToTest"}>
                {todo}
            </div>
        })}
    </>;
}