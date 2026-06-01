'use client'

// TODO: grain batches list and display need fixing
// TODO: plugs list and display need fixing
// TODO: projects list and display need fixing
// TODO: sporeSwab view not working. https://mush.appli.ng/view/sporeSwab/2g1j95Bw5gB
// TODO: transfers list needs fixing. Display appears fine
// TODO: users list needs fixing. Display needs changing?

import {OptionalKey, OptionalSimpleKey} from "@/app/components/common";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {SelectorFor} from "@/app/components/selector";
import {ProjectsSelector, ReadWriteSelector} from "@/app/components/projectClient";
import {UserSelector} from "@/app/components/userClient";
import * as React from "react";
import {useContext, useEffect, useState} from "react";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {RemoveButton} from "@/app/components/formSubcomponents/commonClient";

export function AssertACL(input: any): asserts input is ACL { // TODO: FIX THIS!!!! NEEDS TO DO MAP STUFF PROPERLY!
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['blanketPerm', 'boolean'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {

        if (!OptionalSimpleKey(key, input, expType)) {
            //console.error("failed when validating NON maps!") // TODO: THIS!
            throw new Error('ACL assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    let complexOptionalMapKeys = new Map<string, (v: any) => boolean>([
        ['users', IsStringMapToBool], // TODO: UNSURE IF WORKING
        ['projects', IsStringMapToBool], // TODO: UNSURE IF WORKING
    ])
    for (let [key, validator] of complexOptionalMapKeys) {
        if (!OptionalKey(key, input, validator)) {
            console.error("failed when validating maps!") // TODO: THIS!
            throw new Error('ACL assertion failure: optional array key ' + key + ' was not valid: ' + JSON.stringify(input[key]));
        }
    }
    if (input.users !== null && input.users !== undefined) {
        input.users = new Map<string, boolean>(Object.entries(input.users) as [string, boolean][]) // TODO: UNSURE IF WORKING
    } else {
        input.users = new Map<string, boolean>(); // TODO: will this screw up null (full-write) ACLs?
    }
    if (input.projects !== null && input.projects !== undefined) {
        input.projects = new Map<string, boolean>(Object.entries(input.projects) as [string, boolean][]) // TODO: UNSURE IF WORKING
    } else {
        input.projects = new Map<string, boolean>(); // TODO: will this screw up null (full-write) ACLs?
    }


    return
}

export function MarshalAcl(acl?: ACL): any {
    if (acl === undefined) {
        return undefined
    }
    let out: any = {}
    if (acl.blanketPerm !== undefined) {
        out.blanketPerm = acl.blanketPerm
    }
    if (("users" in acl) && acl.users !== undefined && acl.users.size !== 0) {
        // TODO: why is this occasionally coming back as an object and not a map????? FIX
        if (acl.users instanceof Map) {
            out.users = Object.fromEntries(acl.users)
            console.log("It's a Map");
        } else {
            out.users = acl.users
        }
    }
    if (("projects" in acl) && acl.projects !== undefined && acl.projects.size !== 0) {
        // TODO: why is this occasionally coming back as an object and not a map????? FIX
        if (acl.projects instanceof Map) {
            out.projects = Object.fromEntries(acl.projects)
            console.log("It's a Map");
        } else {
            out.projects = acl.projects
        }

    }
    return out

}

export function IsValidAcl(input: any): boolean {
    try { // TODO: ensure ok
        AssertACL(input) // TODO: may not properly replace the ACL
        return true
    } catch (error) {
        console.error("acl invalid") // TODO: del
        console.error(error) // TODO: del
        return false
    }
}

// TODO: validate ok
export function IsStringMapToString(data: any): data is Record<string, boolean | undefined> {
    // 1. Check if the input is an object and not null.
    if (typeof data !== 'object') {
        return false;
    }

    // 2. Iterate over all keys of the object.
    for (const key in data) {
        if (Object.prototype.hasOwnProperty.call(data, key)) {
            if (typeof data[key] !== 'string') {
                //console.log("typeof entry: " + typeof data[key]) // TODO: delete
                return false;
            }
        }
    }

    // 3. If all values are booleans, it is a map of string to bool.
    return true;
}

export function IsStringMapToBool(data: any): data is Record<string, boolean> {
    // 1. Check if the input is an object and not null.
    if (typeof data !== 'object' || data === null || data === undefined) {
        return false;
    }

    // 2. Iterate over all keys of the object.
    for (const key in data) {
        if (Object.prototype.hasOwnProperty.call(data, key)) {
            // In JavaScript, object keys are always strings.
            // We only need to check the type of each value.
            if (!(typeof data[key] === 'boolean')) {
                // If any value is not a boolean, it fails the check.
                return false;
            }
        }
    }

    // 3. If all values are booleans, it is a map of string to bool.
    return true;
}

export function ProjectsDisplay({readonly, perms, onClick, updateParent, allowAddingCompletedProjects}: {
    readonly: boolean,
    perms: Map<string, boolean>,
    onClick?: (proj: string) => void
    updateParent?: (p: Map<string, boolean>) => void
    allowAddingCompletedProjects?: boolean // TODO: USE THIS IN CALLERS!
}) {
    console.log("current perms: " + JSON.stringify(Object.fromEntries(perms)))
    const projectNameAreaFor = (proj: [string, boolean]) => {
        return <text onClick={() => {
            onClick && onClick(proj[0])
        }}>{proj[0]}</text>
    }
    if (readonly) {
        {
            (perms.size !== 0) && [...perms.entries()].map(values => {
                return <div key={values[0]}>
                    {projectNameAreaFor(values)}
                    <text>{"Can " + (values[1] ? "edit" : "view")}</text>
                </div>
            })
        }
    }
    const update = (newPs: Map<string, boolean>) => {
        //console.log("updating to " + JSON.stringify(Object.fromEntries(newPs)));  // TODO: DEL
        updateParent && updateParent(newPs)
    }
    const addNewProject = (projName: string) => {
        //console.log("adding project: ", projName, " with perms length ", perms.size) // TODO: DEL
        //console.log("before: " + JSON.stringify(Object.fromEntries(perms))); // TODO: DEL
        if (perms.size === 0) {
            const nm = new Map<string, boolean>().set(projName, false)
            //console.log("sending to update for previously empty: " + JSON.stringify(Object.fromEntries(nm))); // TODO: DEL
            update(nm)
        } else {
            const nm = new Map<string, boolean>([...perms.entries()]).set(projName, false)
            //console.log("sending to update for non-empty: " + JSON.stringify(Object.fromEntries(nm))); // TODO: DEL
            update(nm)
        }
        return
    }
    const removeProject = (projName: string) => {
        const newProjs = new Map(perms)
        newProjs.delete(projName)
        update(newProjs)
        return
    }
    return <div>
        <TestAndValidate>
            {(perms === undefined || perms.size === 0) ? null : [...perms.entries()].map(values => { // TODO: THE ENTRIES MAP HERE IS THE CURRENT PROBLEM
                return <div key={values[0]}>
                    {projectNameAreaFor(values)}
                    <SelectorFor options={["can view", "can edit"]} initial={values[1] ? "can edit" : "can view"}
                                 updateParent={s => {
                                     const newProjs = new Map(perms).set(values[0], s === "can edit") // TODO: do this for users too!
                                     update(newProjs)
                                 }} disabled={false}/>
                    <text>{"Can " + (values[1] ? "edit" : "view")}</text>
                    <RemoveButton txt={"Remove project"} click={() => {
                        removeProject(values[0])
                    }}/>
                </div>
            })}
            {"Add a project:"}
            <ProjectsSelector onSelect={addNewProject} complete={allowAddingCompletedProjects}
                              blacklist={(perms.size !== 0) ? [...perms.entries()].map(val => {
                                  return val[0]
                              }) : []}/>
        </TestAndValidate>
    </div>
}

export function AclProjectsDisplay({readonly, ACL, onUsernClick, updateParent}: {
    readonly: boolean,
    ACL?: ACL,
    onUsernClick?: (proj: string) => void // TODO: do we need both onClick and updateParent?
    updateParent?: (acl: ACL) => void
}) {
    return <ProjectsDisplay readonly={readonly}
                            allowAddingCompletedProjects={false/* TODO: ok?*/}
                            perms={(ACL === undefined || ACL.projects === undefined || ACL.projects.size === 0) ? new Map<string, boolean>() : new Map(ACL.projects)}
                            onClick={onUsernClick} updateParent={(projPerms) => {
        let upd = {...ACL}
        upd.projects = projPerms
        console.log("final", JSON.stringify(Object.fromEntries(upd.projects))) // TODO: DEL
        updateParent && updateParent(upd)
    }}/>
}

// TODO: USE THIS!
export function AclUsersDisplayInternal({readonly, perms, onClick, updateParent, blanket}: {
    readonly: boolean,
    perms: Map<string, boolean>,
    onClick?: (usr: string) => void
    updateParent?: (p: Map<string, boolean>) => void
    blanket?: boolean
}) {
    const userNameAreaFor = (val: [string, boolean]) => {
        return <text onClick={(e) => {
            // TODO: stopProp or prevDef?
            // TODO: why val[0]?
            onClick && onClick(val[0])
        }}>{val[0]}</text>
    }
    if (readonly) {
        // TODO: Should have incremented depth?
        {
            perms !== undefined && perms.size > 0 && [...perms.entries()].map(values => {
                return <div key={values[0]}>
                    {userNameAreaFor(values)}
                    <text>{"Can " + (values[1] ? "edit" : "view")}</text>
                </div>
            })
        }
    }
    const update = (perms: Map<string, boolean>) => {
        updateParent && updateParent(perms)
    }
    const addNewUser = (uName: string) => {
        const defaultPerm = blanket || false
        if (perms.size === 0) {/// TODO: ensure ok
            const nm = (new Map<string, boolean>()).set(uName, defaultPerm)
            //console.log("sending to update for previously empty: " + JSON.stringify(Object.fromEntries(nm)));  // TODO: DEL
            update(nm)
        } else {
            const nm = structuredClone(perms).set(uName, defaultPerm)
            //const nm = (new Map<string, boolean>([...perms.entries()])).set(uName, defaultPerm)
            //console.log("sending to update for non-empty: " + JSON.stringify(Object.fromEntries(nm)));  // TODO: DEL
            update(nm)
        }
        return
    }
    const removeUsr = (projName: string) => {
        const newUsrs = new Map(perms)
        newUsrs.delete(projName)
        update(newUsrs)
        return
    }
    const tempFunc = () => {
        //console.log("perms before crash",JSON.stringify(Object.fromEntries(perms))) // TODO: DEL
        return <></>
    }

    //console.log("current perms: "+JSON.stringify(Object.fromEntries(perms))) // TODO: DEL
    return <DepthProvider>
        <div>
            {tempFunc()}

            {(perms !== undefined && perms.size > 0) && [...perms.entries()].map(values => {
                return <div key={values[0]}>
                    {userNameAreaFor(values)}
                    <ReadWriteSelector value={values[1]} readonly={false} onUpdate={b => {
                        const newUsrs = new Map(perms).set(values[0], b)
                        update(newUsrs)
                    }}/>
                    <text>{"Can " + (values[1] ? "edit" : "view")}</text>
                    <RemoveButton txt={"Remove user"} click={() => {
                        removeUsr(values[0])
                    }}/>
                </div>
            })}
            {"Add a user:"}
            <UserSelector onSelect={(u) => {
                addNewUser(u._id)
            }} blacklist={(perms !== undefined && perms.size > 0) ? [...perms.entries()].map(val => {
                return val[0]
            }) : []}/>
        </div>
    </DepthProvider>
}

export function AclBlanketDisplay(inp: {
    readonly: boolean,
    ACL?: ACL,
    updateParent: (a?: boolean) => void
}) {
    // TODO: UPDATE TO STORE ACL LOCALLY, THEN UPDATE PARENT
    const [val, setVal] = useState(inp.ACL)
    useEffect(() => {
        setVal(inp.ACL)
    }, [inp.ACL])
    const updateBlanket = (b?: boolean) => {
        if (val !== undefined) {
            let temp = structuredClone(val)
            temp.blanketPerm = b
            setVal(temp)
        } else {
            setVal({ // TODO: ensure works properly
                users: new Map<string, boolean>(),
                projects: new Map<string, boolean>(),
                blanketPerm: b
            })
        }
        inp.updateParent && inp.updateParent(b)
    }
    if (inp.readonly) {
        return <div>{(val === undefined || val.blanketPerm === true) ? "Publicly Editable" : ((val.blanketPerm === false) ? "Publicly Viewable" : "Private")}</div>
    }
    const permToStr = (a?: ACL) => {
        if (a === undefined || a.blanketPerm === true) {
            return "Publicly Editable"
        } else if (a.blanketPerm === undefined) {
            return "Private"
        }
        return "Publicly Viewable"
    }
    const strToPerm = (s: string) => {
        switch (s) {
            case "Publicly Editable":
                return true
            case "Publicly Viewable":
                return false
            default:
                return undefined
        }
    }
    return <SelectorFor initial={permToStr(inp.ACL)} options={["Private", "Publicly Viewable", "Publicly Editable"]}
                        updateParent={(s) => {
                            inp.updateParent(strToPerm(s))
                        }} disabled={false}/>
}

export function TogglableAreaWithDepth(props: React.PropsWithChildren<{
    startOpen: boolean,
    openTxt?: string,
    closeTxt?: string
}>) {
    const [open, setOpen] = useState(props.startOpen);
    const toggle = () => {
        setOpen(!open);
    }
    if (!open) {
        return <button className={"basicButton"} onClick={toggle}>{props.openTxt || "open"}</button>
    }
    const depth = useContext(DepthContext)
    return <DepthProvider>
        <div className={"subForm depth" + depth}>
            {props.children}
            <button className={"basicButton"} onClick={toggle}>{props.closeTxt || "close"}</button>
        </div>
    </DepthProvider>
}

export function CloneAcl(ACL?: ACL): ACL { // TODO: use?
    if (ACL === undefined) {
        return {users: new Map<string, boolean>(), projects: new Map<string, boolean>()}
    }
    let out: ACL = {users: new Map<string, boolean>(), projects: new Map<string, boolean>()}
    if (ACL.users !== undefined && ACL.users.size !== 0) {
        ACL.users = new Map<string, boolean>(ACL.users);
    }
    if (ACL.projects !== undefined && ACL.projects.size !== 0) {
        ACL.projects = new Map<string, boolean>(ACL.projects);
    }
    out.blanketPerm = ACL.blanketPerm;
    return out;
}


export function AclDefaultAclDisplay(inp: {
    readonly: boolean,
    ACL?: ACL,
    defaultACL?: ACL,
    updateAcl: (acl?: ACL) => void
    updateDefaultAcl: (acl?: ACL) => void
}) {
    const depth = useContext(DepthContext)
    const [open, setOpen] = useState<string | undefined>(undefined)
    const hideButton = <button className={"basicButtonSmall"}>{"Hide ACL"}</button>
    const aclArea = () => {
        if (!open) {
            return null
        }
        if (open === "ACL") {
            return <><AclDisplay ACL={inp.ACL} updateParent={inp.updateAcl} readonly={inp.readonly}/>
                {hideButton}
            </>
        }
        return <><AclDisplay ACL={inp.defaultACL} updateParent={inp.updateDefaultAcl} readonly={inp.readonly}/>
            {hideButton}
        </>
    }
    const btnClicked = (v: string) => {
        setOpen((open && v === open) ? undefined : v)
    }
    const aButton = <div className={"depth" + (open === "ACL" ? (depth) + " closeable" : (depth + 1) + " activatable")}
                         onClick={() => {
                             btnClicked("ACL")
                         }}>{"ACL"}</div>
    const bButton = <div
        className={"depth" + (open === "defaultACL" ? (depth) + " closeable" : (depth + 1) + " activatable")}
        onClick={() => {
            btnClicked("defaultACL")
        }}>{"Default ACL"}</div>
    return <div className={"subForm depth" + depth}>
        <div className={"aclDualButton"}>
            {aButton}
            {bButton}
        </div>
        {aclArea()}
    </div>
}

export function AclDisplay(inp: {
    readonly: boolean,
    //initial?: ACL,
    ACL?: ACL,
    updateParent: (acl?: ACL) => void
}) {
    // const [current, setCurrent] = useState(inp.initial)
    // useEffect(()=>{
    //     setCurrent(inp.initial)
    // },[inp.initial])
    const depth = useContext(DepthContext)
    if (inp.readonly) {
        return <div>{/* TODO: TURN INTO A TABLE!!!!*/}
            <TestAndValidate todos={["preloaded values arent sticking around when updating.... troubleshoot"]}>
                <AclBlanketDisplay readonly={true} ACL={inp.ACL} updateParent={(bp) => {
                    let users = inp.ACL ?
                        ((inp.ACL.users !== undefined && inp.ACL.users.size !== 0) ? (new Map<string, boolean>(inp.ACL.users)) : new Map<string, boolean>()) :
                        new Map<string, boolean>()
                    let projects = inp.ACL ?
                        ((inp.ACL.projects !== undefined && inp.ACL.projects.size !== 0) ? (new Map<string, boolean>(inp.ACL.projects)) : new Map<string, boolean>()) :
                        new Map<string, boolean>()
                    inp.updateParent && inp.updateParent({projects: projects, users: users, blanketPerm: bp})
                }}/>
                <AclUsersDisplayInternal readonly={true}
                                         perms={(inp.ACL !== undefined && inp.ACL.users !== undefined) ? inp.ACL.users : (new Map<string, boolean>())}
                                         onClick={() => {/* TODO: onClick?*/
                                         }} updateParent={u => {
                    inp.updateParent({...(inp.ACL), users: u})
                }}/>
                <AclProjectsDisplay readonly={true} ACL={inp.ACL} onUsernClick={() => {/* TODO: onClick?*/
                }} updateParent={inp.updateParent}/>
            </TestAndValidate>
        </div>

    }
    return <div className={"subForm depth" + depth}>
        <AclBlanketDisplay readonly={false} ACL={inp.ACL} updateParent={(b?: boolean) => {
            inp.updateParent({...(inp.ACL), blanketPerm: b})
        }}/>
        <AclUsersDisplayInternal readonly={false} blanket={inp.ACL?.blanketPerm}
                                 perms={(inp.ACL !== undefined && inp.ACL.users !== undefined) ? inp.ACL.users : (new Map<string, boolean>())}
                                 onClick={() => {
                                 }} updateParent={(us) => {
            inp.updateParent({...structuredClone(inp.ACL), users: us})
        }}/>
        <AclProjectsDisplay readonly={false} ACL={inp.ACL} onUsernClick={() => {
        }} updateParent={(newAcl) => {
            inp.updateParent(structuredClone(newAcl))
        }}/>
    </div>
}

export function DefaultAclDisplay({readonly, ACL, updateParent}: {
    readonly: boolean,
    ACL?: ACL,
    updateParent: (acl?: ACL) => void
}) {
    return <div>
        {"Default Access Control List"}
        <AclDisplay readonly={readonly} ACL={ACL} updateParent={updateParent}/>
    </div>
}